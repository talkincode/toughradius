package radiusd

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
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

	s := unsafe.Sizeof(header)
	var data = make([]byte, header.Length-uint16(s))
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
