package tlsengine

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"
)

// establishTunnel completes a server-only TLS handshake against eng while
// keeping the real tls.Client alive afterwards, so a test can exchange
// application data and export keying material over the established tunnel. It
// mirrors driveHandshake but does not close the client.
func establishTunnel(t *testing.T, eng *Engine, clientCfg *tls.Config) (client *tls.Conn, toClient, fromClient *bufStream) {
	t.Helper()
	clientConn, toClient, fromClient := newClientPipe()
	client = tls.Client(clientConn, clientCfg)

	clientDone := make(chan error, 1)
	go func() { clientDone <- client.Handshake() }()

	buf := make([]byte, 16*1024)
	in := []byte(nil)
	deadline := time.After(5 * time.Second)
	for {
		out, done, err := eng.Process(in)
		if err != nil {
			t.Fatalf("server handshake: %v", err)
		}
		if len(out) > 0 {
			if _, werr := toClient.Write(out); werr != nil {
				t.Fatalf("write to client: %v", werr)
			}
		}
		if done {
			break
		}
		select {
		case <-deadline:
			t.Fatal("server handshake timed out")
		default:
		}
		n, rerr := fromClient.Read(buf)
		if rerr != nil {
			t.Fatalf("read client flight: %v", rerr)
		}
		in = append([]byte(nil), buf[:n]...)
	}

	select {
	case err := <-clientDone:
		if err != nil {
			t.Fatalf("client handshake: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("client handshake did not return")
	}
	return client, toClient, fromClient
}

func serverOnlyEngine(t *testing.T, ca *testCA, minVersion uint16) *Engine {
	t.Helper()
	serverCert := ca.issue(t, "radius.example.com", func(c *x509.Certificate) {
		c.DNSNames = []string{"radius.example.com"}
	})
	eng, err := New(&Config{
		ServerCertificate: serverCert,
		ServerOnly:        true,
		MinVersion:        minVersion,
		HandshakeTimeout:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New server-only engine: %v", err)
	}
	return eng
}

func serverOnlyClientCfg(ca *testCA, maxVersion uint16) *tls.Config {
	return &tls.Config{
		RootCAs:    ca.pool,
		ServerName: "radius.example.com",
		MinVersion: tls.VersionTLS12,
		MaxVersion: maxVersion,
	}
}

// TestEngine_ApplicationDataRoundTrip verifies that, after a server-only
// handshake, the engine can send and receive application data in both
// directions over the established tunnel — the foundation PEAP/TTLS use to carry
// an inner EAP method.
func TestEngine_ApplicationDataRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name       string
		maxVersion uint16
	}{
		{"TLS1.2", tls.VersionTLS12},
		{"TLS1.3", tls.VersionTLS13},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ca := newTestCA(t, "AppData Root CA")
			eng := serverOnlyEngine(t, ca, tls.VersionTLS12)
			defer func() { _ = eng.Close() }()

			client, toClient, fromClient := establishTunnel(t, eng, serverOnlyClientCfg(ca, tc.maxVersion))
			defer func() { _ = client.Close() }()

			// Server -> client: the engine encrypts an inner request and the
			// real client decrypts it.
			request := []byte("inner-eap-request")
			records, err := eng.WriteApplication(request)
			if err != nil {
				t.Fatalf("WriteApplication: %v", err)
			}
			if len(records) == 0 {
				t.Fatal("WriteApplication produced no records")
			}
			if _, err := toClient.Write(records); err != nil {
				t.Fatalf("deliver records to client: %v", err)
			}
			buf := make([]byte, 4096)
			n, err := client.Read(buf)
			if err != nil {
				t.Fatalf("client.Read: %v", err)
			}
			if !bytes.Equal(buf[:n], request) {
				t.Fatalf("client decrypted %q, want %q", buf[:n], request)
			}

			// Client -> server: the real client encrypts an inner response and
			// the engine decrypts it.
			response := []byte("inner-eap-response")
			if _, err := client.Write(response); err != nil {
				t.Fatalf("client.Write: %v", err)
			}
			rbuf := make([]byte, 4096)
			rn, err := fromClient.Read(rbuf)
			if err != nil {
				t.Fatalf("read client record: %v", err)
			}
			got, err := eng.ReadApplication(rbuf[:rn])
			if err != nil {
				t.Fatalf("ReadApplication: %v", err)
			}
			if !bytes.Equal(got, response) {
				t.Fatalf("server decrypted %q, want %q", got, response)
			}
		})
	}
}

// TestEngine_ExportKey verifies that exported keying material derived on the
// server engine matches the value the peer derives, and has the requested
// length — the property PEAP/TTLS rely on to derive identical MS-MPPE keys.
func TestEngine_ExportKey(t *testing.T) {
	for _, tc := range []struct {
		name       string
		maxVersion uint16
	}{
		{"TLS1.2", tls.VersionTLS12},
		{"TLS1.3", tls.VersionTLS13},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ca := newTestCA(t, "ExportKey Root CA")
			eng := serverOnlyEngine(t, ca, tls.VersionTLS12)
			defer func() { _ = eng.Close() }()

			client, _, _ := establishTunnel(t, eng, serverOnlyClientCfg(ca, tc.maxVersion))
			defer func() { _ = client.Close() }()

			const label = "client EAP encryption"
			serverKey, err := eng.ExportKey(label, nil, 64)
			if err != nil {
				t.Fatalf("server ExportKey: %v", err)
			}
			if len(serverKey) != 64 {
				t.Fatalf("server key length = %d, want 64", len(serverKey))
			}

			cs := client.ConnectionState()
			clientKey, err := cs.ExportKeyingMaterial(label, nil, 64)
			if err != nil {
				t.Fatalf("client ExportKeyingMaterial: %v", err)
			}
			if !bytes.Equal(serverKey, clientKey) {
				t.Fatalf("server/client exported keys differ:\n server=%x\n client=%x", serverKey, clientKey)
			}
		})
	}
}

// TestEngine_AppMethods_BeforeHandshake ensures the post-handshake methods
// refuse to operate until the handshake has completed, so a tunneled method can
// never read, write, or export keys from an unestablished tunnel.
func TestEngine_AppMethods_BeforeHandshake(t *testing.T) {
	ca := newTestCA(t, "PreHandshake Root CA")
	eng := serverOnlyEngine(t, ca, tls.VersionTLS12)
	defer func() { _ = eng.Close() }()

	if eng.HandshakeComplete() {
		t.Fatal("HandshakeComplete should be false before the handshake")
	}
	if _, err := eng.WriteApplication([]byte("x")); err != ErrHandshakeNotComplete {
		t.Fatalf("WriteApplication before handshake = %v, want ErrHandshakeNotComplete", err)
	}
	if _, err := eng.ReadApplication([]byte("x")); err != ErrHandshakeNotComplete {
		t.Fatalf("ReadApplication before handshake = %v, want ErrHandshakeNotComplete", err)
	}
	if _, err := eng.ExportKey("client EAP encryption", nil, 64); err != ErrHandshakeNotComplete {
		t.Fatalf("ExportKey before handshake = %v, want ErrHandshakeNotComplete", err)
	}
}

// TestEngine_ReadApplication_Timeout verifies that an incomplete inbound TLS
// record causes ReadApplication to time out (rather than block a RADIUS worker
// goroutine forever) and that the engine is left closed.
func TestEngine_ReadApplication_Timeout(t *testing.T) {
	ca := newTestCA(t, "Timeout Root CA")
	eng := serverOnlyEngine(t, ca, tls.VersionTLS12)
	defer func() { _ = eng.Close() }()

	client, _, _ := establishTunnel(t, eng, serverOnlyClientCfg(ca, tls.VersionTLS13))
	defer func() { _ = client.Close() }()

	eng.appReadTimeout = 150 * time.Millisecond

	// A single byte is not a complete TLS record, so the decrypt read cannot
	// make progress and must hit the timeout.
	if _, err := eng.ReadApplication([]byte{0x17}); err != ErrAppReadTimeout {
		t.Fatalf("ReadApplication with partial record = %v, want ErrAppReadTimeout", err)
	}
}
