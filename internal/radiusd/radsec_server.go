package radiusd

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"layeh.com/radius"
)

// defaultRadsecWorkers bounds the number of concurrent RadSec handler goroutines
// when RadsecWorker is left unconfigured (<= 0). A zero-capacity worker channel
// would otherwise be unbuffered and deadlock the read loop, so a sane default is
// applied. It mirrors config.DefaultAppConfig.Radiusd.RadsecWorker.
const defaultRadsecWorkers = 100

// maxRadiusPacketLength is the maximum RADIUS packet length per RFC 2865 §3. The
// 2-byte Length field can encode up to 65535, but a conformant packet never
// exceeds 4096 bytes; rejecting anything larger bounds per-packet allocations
// (and the size a pooled parse buffer can retain).
const maxRadiusPacketLength = 4096

// errMalformedFrame indicates the RadSec length prefix could not be trusted to
// delimit the frame: it is smaller than a valid packet or larger than the RFC
// 2865 maximum. Because the announced body bytes are NOT consumed from the
// stream when this is returned, the TCP connection is desynchronized — the
// remaining bytes would be misread as subsequent packet headers. Callers must
// treat it as fatal and close the connection rather than continue parsing.
var errMalformedFrame = errors.New("radius: malformed packet framing")

type packetResponseWriter struct {
	// listener that received the packet
	conn net.Conn
	addr net.Addr
}

type RadsecHandler interface {
	ServeRADIUS(w radius.ResponseWriter, r *radius.Request)
}

func (r *packetResponseWriter) Write(packet *radius.Packet) error {
	encoded, err := packet.Encode()
	if err != nil {
		return err
	}
	if _, err := r.conn.Write(encoded); err != nil {
		return err
	}
	return nil
}

// RadsecPacketServer listens for RADIUS requests on a packet-based protocols (e.g.
// UDP).
type RadsecPacketServer struct {
	// The address on which the server listens. Defaults to :1812.
	Addr string

	// The network on which the server listens. Defaults to udp.
	Network string

	// The source from which the secret is obtained for parsing and validating
	// the request.
	SecretSource radius.SecretSource

	// Handler which is called to process the request.
	Handler RadsecHandler

	// Skip incoming packet authenticity validation.
	// This should only be set to true for debugging purposes.
	InsecureSkipVerify bool

	RadsecWorker int

	// ErrorLog specifies an optional logger for errors
	// around packet accepting, processing, and validation.
	// If nil, logging is done via the log package's standard logger.
	// ErrorLog *log.Logger

	shutdownRequested int32

	mu          sync.Mutex
	ctx         context.Context
	ctxDone     context.CancelFunc
	listeners   map[net.Conn]uint
	lastActive  chan struct{} // closed when the last active item finishes
	activeCount int32
	workerPool  chan struct{}
}

func (s *RadsecPacketServer) initLocked() {
	if s.ctx == nil {
		workers := s.RadsecWorker
		if workers <= 0 {
			workers = defaultRadsecWorkers
		}
		s.ctx, s.ctxDone = context.WithCancel(context.Background())
		s.listeners = make(map[net.Conn]uint)
		s.lastActive = make(chan struct{})
		s.workerPool = make(chan struct{}, workers)
	}
}

// acquireWorkerSlot reserves a slot in the bounded worker pool before a handler
// goroutine is spawned. It returns false when the server is shutting down.
//
// Acquiring the slot synchronously in the read loop (rather than inside the
// spawned goroutine) is what actually bounds the number of in-flight handler
// goroutines to the pool capacity: when the pool is saturated the read loop
// stops pulling packets off the connection, which applies natural TCP back-
// pressure instead of letting a flood spawn unbounded goroutines. This mirrors
// the bounded back-pressure used on the UDP accounting path
// (AcctService.submitAcctTask), adapted to RadSec's connection-oriented transport.
func (s *RadsecPacketServer) acquireWorkerSlot() bool {
	// Honor shutdown first: once the server is stopping, never start a new
	// handler even if the pool still has free slots. This keeps the documented
	// "returns false when shutting down" contract consistent.
	if s.ctx.Err() != nil {
		return false
	}
	select {
	case s.workerPool <- struct{}{}:
		return s.keepSlotUnlessShutdown()
	default:
	}
	// Pool saturated: record the back-pressure event for observability, then
	// block until a slot frees or the server shuts down.
	app.IncRadiusMetric(app.MetricsRadiusRadsecSaturated)
	select {
	case s.workerPool <- struct{}{}:
		return s.keepSlotUnlessShutdown()
	case <-s.ctx.Done():
		return false
	}
}

// keepSlotUnlessShutdown re-checks the shutdown signal after a slot was acquired.
// Acquiring a free slot (a non-blocking channel send) and observing ctx are two
// separate steps, so Shutdown can cancel ctx in between. Re-checking here and
// releasing the slot on shutdown closes that race deterministically: no new
// handler is started once shutdown has begun, and the slot is handed back.
func (s *RadsecPacketServer) keepSlotUnlessShutdown() bool {
	if s.ctx.Err() != nil {
		s.releaseWorkerSlot()
		return false
	}
	return true
}

// releaseWorkerSlot returns a slot to the bounded worker pool. It must be called
// exactly once for every successful acquireWorkerSlot.
func (s *RadsecPacketServer) releaseWorkerSlot() {
	<-s.workerPool
}

func (s *RadsecPacketServer) activeAdd() {
	atomic.AddInt32(&s.activeCount, 1)
}

func (s *RadsecPacketServer) activeDone() {
	if atomic.AddInt32(&s.activeCount, -1) == -1 {
		close(s.lastActive)
	}
}

// tcpPacketBufferPool recycles the per-packet payload buffer used by
// parseTcpPacket. radius.ParseAttributes copies attribute bytes out (it appends
// into freshly allocated slices rather than aliasing the input), and the request
// authenticator is copied into a fixed-size array, so the buffer holds no live
// references once parseTcpPacket returns and can be safely reused across packets.
//
// Pointers to slices are pooled (not slices directly) to avoid allocating a slice
// header on every Put. pprof identified this buffer as the largest allocation in
// our own RadSec ingest path; the remaining per-packet allocations live inside
// layeh.com/radius and are out of scope for in-tree pooling.
var tcpPacketBufferPool = sync.Pool{
	New: func() any {
		// Most RADIUS packets are well under 512 bytes; the buffer grows on demand.
		b := make([]byte, 0, 512)
		return &b
	},
}

func parseTcpPacket(r io.Reader, secret []byte) (*radius.Packet, error) {
	var header struct {
		Code       uint8
		Identifier uint8
		Length     uint16
	}

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	headerSize := uint16(unsafe.Sizeof(header))
	// Guard the length field before allocating or reading. A Length below the
	// header size would underflow the unsigned subtraction; a payload shorter
	// than the 16-byte authenticator would make data[16:] panic; and a Length
	// above the RFC 2865 maximum (4096) must be rejected so a malicious/garbled
	// length cannot drive a large allocation+read or, via the buffer pool,
	// permanently retain an oversized backing array. These are framing errors:
	// the body has not been consumed, so the caller must close the connection.
	if header.Length < headerSize || header.Length > maxRadiusPacketLength {
		return nil, fmt.Errorf("%w: length %d", errMalformedFrame, header.Length)
	}
	dataLen := int(header.Length - headerSize)
	if dataLen < 16 {
		return nil, fmt.Errorf("%w: payload shorter than authenticator", errMalformedFrame)
	}

	bufp := tcpPacketBufferPool.Get().(*[]byte)
	defer tcpPacketBufferPool.Put(bufp)
	if cap(*bufp) < dataLen {
		*bufp = make([]byte, dataLen)
	}
	data := (*bufp)[:dataLen]

	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	attrs, err := radius.ParseAttributes(data[16:])
	if err != nil {
		return nil, err
	}

	packet := &radius.Packet{
		Code:       radius.Code(header.Code),
		Identifier: header.Identifier,
		Secret:     secret,
		Attributes: attrs,
	}
	copy(packet.Authenticator[:], data[0:16])
	return packet, nil
}

// Serve accepts incoming connections on conn.
func (s *RadsecPacketServer) Serve(conn net.Conn) error {
	// Always close the connection when Serve returns, on every exit path
	// (nil handler/secret, EOF, network error, fatal framing error, or
	// shutdown). ListenAndServe runs Serve in a goroutine and discards its
	// return value, so without this the underlying socket would stay open
	// until GC/finalizers ran, leaking connections under attack or noisy
	// clients. Close is idempotent-safe here: Shutdown may also close this
	// conn via s.listeners, and net.Conn permits concurrent Close/Write, so an
	// in-flight handler writing to an already-terminating connection simply
	// gets a write error rather than racing.
	defer func() { _ = conn.Close() }() //nolint:errcheck

	if s.Handler == nil {
		return errors.New("radius: nil RadsecHandler")
	}
	if s.SecretSource == nil {
		return errors.New("radius: nil SecretSource")
	}

	s.mu.Lock()
	s.initLocked()
	if atomic.LoadInt32(&s.shutdownRequested) == 1 {
		s.mu.Unlock()
		return radius.ErrServerShutdown
	}

	s.listeners[conn]++
	s.mu.Unlock()

	type requestKey struct {
		IP         string
		Identifier byte
	}

	var (
		requestsLock sync.Mutex
		requests     = map[requestKey]struct{}{}
	)

	s.activeAdd()
	defer func() {
		s.mu.Lock()
		s.listeners[conn]--
		if s.listeners[conn] == 0 {
			delete(s.listeners, conn)
		}
		s.mu.Unlock()
		s.activeDone()
	}()

	secret, err := s.SecretSource.RADIUSSecret(s.ctx, conn.RemoteAddr())
	if err != nil {
		zap.S().Errorf("radius: error fetching from secret source: %v", err)
		return err
	}

	if len(secret) == 0 {
		zap.S().Errorf("radius: empty secret returned from secret source")
		return err
	}

	r := bufio.NewReader(conn)

	for {
		pkt, err := parseTcpPacket(r, secret)
		if err != nil {
			if err == io.EOF {
				zap.S().Infof("radius: connection closed by client %s", conn.RemoteAddr())
				return err
			}
			if _, ok := err.(net.Error); ok {
				zap.S().Infof("radius: connection error %s: %v", conn.RemoteAddr(), err)
				return err
			}
			// A framing error means the length prefix could not delimit the
			// frame and its announced body was not consumed, so the stream is
			// desynchronized. Continuing would misread the body as the next
			// header; close the connection instead. The client may reconnect.
			if errors.Is(err, errMalformedFrame) {
				zap.S().Errorf("radius: closing connection %s on malformed frame: %v", conn.RemoteAddr(), err)
				return err
			}
			zap.S().Errorf("radius: unable to parse packet: %v", err)
			continue
		}

		// Reserve a worker slot before spawning the handler goroutine. This
		// bounds in-flight handler goroutines to the pool size and applies TCP
		// back-pressure when saturated instead of spawning goroutines without
		// limit. On shutdown the acquire is interrupted and Serve returns.
		if !s.acquireWorkerSlot() {
			return radius.ErrServerShutdown
		}

		s.activeAdd()
		go func(packet *radius.Packet, conn net.Conn) {
			defer s.activeDone()
			defer s.releaseWorkerSlot()

			key := requestKey{
				IP:         conn.RemoteAddr().String(),
				Identifier: packet.Identifier,
			}

			requestsLock.Lock()
			if _, ok := requests[key]; ok {
				requestsLock.Unlock()
				return
			}
			requests[key] = struct{}{}
			requestsLock.Unlock()

			response := packetResponseWriter{
				conn: conn,
				addr: conn.RemoteAddr(),
			}

			defer func() {
				requestsLock.Lock()
				delete(requests, key)
				requestsLock.Unlock()
			}()

			request := radius.Request{
				LocalAddr:  conn.LocalAddr(),
				RemoteAddr: conn.RemoteAddr(),
				Packet:     packet,
			}

			s.Handler.ServeRADIUS(&response, &request)
		}(pkt, conn)
	}
}

// initTLSConfig initializes a tls.Config with the given certificate and key
func (s *RadsecPacketServer) initTLSConfig(capath, crtfile, keyfile string) (*tls.Config, error) {
	crt, err := tls.LoadX509KeyPair(crtfile, keyfile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{crt},
		Time:         time.Now,
		Rand:         rand.Reader,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		MinVersion:   tls.VersionTLS12,
	}

	if common.FileExists(capath) {
		cabytes, _ := os.ReadFile(capath) //nolint:gosec // G304: path is from validated config
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(cabytes)
		tlsConfig.ClientCAs = pool
	}

	return tlsConfig, nil
}

// ListenAndServe starts a RADIUS server on the address given in s.
func (s *RadsecPacketServer) ListenAndServe(capath, crtfile, keyfile string) error {
	tlsConfig, err := s.initTLSConfig(capath, crtfile, keyfile)
	if err != nil {
		return err
	}

	if s.Handler == nil {
		return errors.New("radius: nil RadsecHandler")
	}
	if s.SecretSource == nil {
		return errors.New("radius: nil SecretSource")
	}

	addrStr := s.Addr
	if addrStr == "" {
		addrStr = ":2083" // Default RadSec port
	}

	network := "tcp"
	if s.Network != "" {
		network = s.Network
	}

	pc, err := tls.Listen(network, addrStr, tlsConfig)
	if err != nil {
		return err
	}
	defer func() { _ = pc.Close() }() //nolint:errcheck
	for {
		conn, err := pc.Accept()
		if err != nil {
			continue
		}
		go s.Serve(conn) //nolint:errcheck
	}
}

// Shutdown gracefully stops the server. It first closes all listeners and then
// waits for any running handlers to complete.
//
// Shutdown returns after nil all handlers have completed. ctx.Err() is
// returned if ctx is canceled.
//
// Any Serve methods return ErrShutdown after Shutdown is called.
func (s *RadsecPacketServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.initLocked()
	if atomic.CompareAndSwapInt32(&s.shutdownRequested, 0, 1) {
		for listener := range s.listeners {
			_ = listener.Close() //nolint:errcheck
		}

		s.ctxDone()
		s.activeDone()
	}
	s.mu.Unlock()

	select {
	case <-s.lastActive:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
