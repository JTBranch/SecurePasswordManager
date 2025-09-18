package lan

import (
	"bytes"
	"context"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	mrand "math/rand"
	"net"
	"sync"
	"time"

	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	transportpb "go-password-manager/internal/transport/proto"

	"github.com/grandcat/zeroconf"
	"google.golang.org/protobuf/proto"
)

// Config defines LAN transport configuration.
type Config struct {
	ListenAddr       string
	DialTimeout      time.Duration
	HandshakeTimeout time.Duration
	MaxSendAttempts  int
	AckWait          time.Duration
	Discovery        bool // advertise via mDNS when true
}

// Transport implements a skeleton LAN BundleTransport. Real message framing
// & crypto handshake will be added in subsequent iterations.
type Transport struct {
	cfg        Config
	local      transport.DeviceDescriptor
	deps       transport.Dependencies
	ln         net.Listener
	closed     chan struct{}
	tlsCert    tls.Certificate
	tlsConfig  *tls.Config
	inbound    chan *transport.InboundBundle
	advertiser interface{ Close() }
	pinned     sync.Map // deviceID -> DER SubjectPublicKeyInfo
}

// Addr returns the listen address of the transport (for tests / diagnostics).
func (t *Transport) Addr() string {
	if t.ln == nil {
		return ""
	}
	return t.ln.Addr().String()
}

func (t *Transport) ID() string { return "lan" }

func (t *Transport) Send(ctx context.Context, bundle *sharing.SecretExportBundle, target transport.DeviceDescriptor) (*transport.TransportReceipt, error) {
	r := &transport.TransportReceipt{TransportID: t.ID(), Target: target, BundleID: "", StartedAt: time.Now(), Attempt: 0, Meta: map[string]any{}}
	if bundle != nil {
		r.BundleID = bundle.Payload.ID
	}
	if bundle == nil {
		r.Error = errors.New("lan: nil bundle")
		r.CompletedAt = time.Now()
		return r, r.Error
	}
	addr := target.LastAddr
	if addr == "" {
		r.Error = errors.New("lan: target address unknown")
		r.CompletedAt = time.Now()
		return r, r.Error
	}
	attempts := t.cfg.MaxSendAttempts
	if attempts <= 0 {
		attempts = 1
	}
	backoff := 50 * time.Millisecond
	for i := 1; i <= attempts; i++ {
		r.Attempt = i
		if err := t.sendAttempt(ctx, addr, bundle); err == nil {
			r.CompletedAt = time.Now()
			return r, nil
		} else {
			r.Error = err
			if i < attempts {
				select {
				case <-ctx.Done():
					r.Error = ctx.Err()
					r.CompletedAt = time.Now()
					return r, r.Error
				case <-time.After(backoff):
				}
				// apply jitter (80% - 120%) then exponential growth
				jitterFactor := 0.8 + mrand.Float64()*0.4
				backoff = time.Duration(float64(backoff)*jitterFactor) * 2
				continue
			}
			r.CompletedAt = time.Now()
			return r, r.Error
		}
	}
	r.CompletedAt = time.Now()
	return r, r.Error
}

func (t *Transport) sendAttempt(ctx context.Context, addr string, bundle *sharing.SecretExportBundle) error {
	d := &net.Dialer{Timeout: t.cfg.DialTimeout}
	conn, err := tls.DialWithDialer(d, "tcp", addr, t.clientTLSConfig())
	if err != nil {
		return err
	}
	defer conn.Close()
	penv := &transportpb.Envelope{Type: transportpb.MessageType_MESSAGE_TYPE_DATA, SenderId: t.local.DeviceID, Bundle: toProtoBundle(bundle)}
	data, err := proto.Marshal(penv)
	if err != nil {
		return err
	}
	if err := writeLenPrefixed(conn, data); err != nil {
		return err
	}
	if t.cfg.AckWait > 0 {
		return t.waitAck(conn)
	}
	return nil
}

func (t *Transport) waitAck(conn net.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(t.cfg.AckWait))
	payload, err := readLenPrefixed(conn)
	if err != nil {
		return err
	}
	var ack transportpb.Envelope
	if err := proto.Unmarshal(payload, &ack); err != nil {
		return err
	}
	if ack.Type == transportpb.MessageType_MESSAGE_TYPE_ERROR {
		return errors.New("lan: remote error")
	}
	if ack.Type != transportpb.MessageType_MESSAGE_TYPE_ACK {
		return errors.New("lan: missing ack")
	}
	return nil
}

func (t *Transport) Receive(ctx context.Context) (*transport.InboundBundle, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ib, ok := <-t.inbound:
		if !ok {
			return nil, errors.New("lan: transport closed")
		}
		return ib, nil
	}
}

func (t *Transport) Close() error {
	select {
	case <-t.closed:
		return nil
	default:
	}
	close(t.closed)
	if t.ln != nil {
		_ = t.ln.Close()
	}
	return nil
}

type factory struct{}

func (factory) New(ctx context.Context, cfg map[string]any, local transport.DeviceDescriptor, deps transport.Dependencies) (transport.BundleTransport, error) {
	c := parseConfig(cfg)
	ln, err := net.Listen("tcp", c.ListenAddr)
	if err != nil {
		return nil, err
	}
	updatedLocal, cert, tlsConf, err := buildTLSIdentity(local, deps)
	if err != nil {
		return nil, err
	}
	t := &Transport{cfg: c, local: updatedLocal, deps: deps, ln: ln, closed: make(chan struct{}), tlsCert: cert, tlsConfig: tlsConf, inbound: make(chan *transport.InboundBundle, 8)}
	if c.Discovery {
		if deps.Advertisement != nil {
			if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
				if handle, err := deps.Advertisement.Start(&t.local, tcpAddr.IP.String(), tcpAddr.Port, nil); err == nil {
					t.advertiser = &advertAdapter{h: handle}
				}
			}
		} else if adv, err := startAdvertisement(&t.local, ln.Addr()); err == nil {
			t.advertiser = zeroconfWrapper{Server: adv}
		}
	}
	go t.acceptLoop()
	return t, nil
}

func parseConfig(cfg map[string]any) Config {
	c := Config{ListenAddr: ":0", DialTimeout: 3 * time.Second, HandshakeTimeout: 5 * time.Second, MaxSendAttempts: 3, AckWait: 750 * time.Millisecond}
	if v, ok := cfg["listen_addr"].(string); ok && v != "" {
		c.ListenAddr = v
	}
	if v, ok := cfg["max_attempts"].(int); ok && v > 0 {
		c.MaxSendAttempts = v
	}
	if v, ok := cfg["ack_wait_ms"].(int); ok && v > 0 {
		c.AckWait = time.Duration(v) * time.Millisecond
	}
	if v, ok := cfg["discovery"].(bool); ok {
		c.Discovery = v
	}
	return c
}

func buildTLSIdentity(local transport.DeviceDescriptor, deps transport.Dependencies) (transport.DeviceDescriptor, tls.Certificate, *tls.Config, error) {
	// Prefer explicit KeyGen dependency; fall back to ephemeral cert.
	if deps.KeyGen != nil && len(local.Ed25519Pub) == 0 {
		pub, priv, err := deps.KeyGen.GenerateEd25519KeyPair()
		if err == nil && len(pub) == ed25519.PublicKeySize && len(priv) == ed25519.PrivateKeySize {
			local.Ed25519Pub = pub
			if cert, conf, err2 := generateTLSConfigFromKey(local, ed25519.PrivateKey(priv)); err2 == nil {
				return local, cert, conf, nil
			}
		}
	}
	cert, conf, err := generateEphemeralTLSConfig(local)
	return local, cert, conf, err
}

func (t *Transport) acceptLoop() {
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			select {
			case <-t.closed:
				return
			default:
			}
			continue
		}
		go t.handleConn(conn)
	}
}

func (t *Transport) handleConn(raw net.Conn) {
	defer raw.Close()
	tlsConn := tls.Server(raw, t.serverTLSConfig())
	if err := tlsConn.Handshake(); err != nil {
		return
	}
	env, ok := t.readEnvelope(tlsConn)
	if !ok {
		return
	}
	if !t.verifyAndPin(tlsConn, env) {
		return
	}
	t.deliverBundle(raw, tlsConn, env)
}

func (t *Transport) readEnvelope(tlsConn net.Conn) (*transportpb.Envelope, bool) {
	payload, err := readLenPrefixed(tlsConn)
	if err != nil {
		return nil, false
	}
	var env transportpb.Envelope
	if err := proto.Unmarshal(payload, &env); err != nil {
		return nil, false
	}
	if env.Type != transportpb.MessageType_MESSAGE_TYPE_DATA || env.Bundle == nil {
		t.sendErrorIfNeeded(tlsConn)
		return nil, false
	}
	return &env, true
}

func (t *Transport) verifyAndPin(tlsConn net.Conn, env *transportpb.Envelope) bool {
	if peer := tlsConn.(*tls.Conn).ConnectionState().PeerCertificates; len(peer) > 0 {
		key := peer[0].RawSubjectPublicKeyInfo
		if existing, ok := t.pinned.Load(env.SenderId); ok {
			if !bytes.Equal(existing.([]byte), key) {
				t.sendErrorIfNeeded(tlsConn)
				return false
			}
		} else {
			t.pinned.Store(env.SenderId, append([]byte(nil), key...))
		}
	}
	return true
}

func (t *Transport) deliverBundle(raw net.Conn, tlsConn net.Conn, env *transportpb.Envelope) {
	bundle := fromProtoBundle(env.Bundle)
	ib := &transport.InboundBundle{From: transport.DeviceDescriptor{DeviceID: env.SenderId}, Bundle: bundle, ReceivedAt: time.Now(), Envelope: map[string]any{"lan_addr": raw.RemoteAddr().String()}}
	select {
	case t.inbound <- ib:
	default:
	}
	if t.cfg.AckWait > 0 {
		t.sendAck(tlsConn)
	}
}

func (t *Transport) sendAck(conn net.Conn) {
	ack := &transportpb.Envelope{Type: transportpb.MessageType_MESSAGE_TYPE_ACK}
	if data, err := proto.Marshal(ack); err == nil {
		_ = writeLenPrefixed(conn, data)
	}
}

func (t *Transport) sendErrorIfNeeded(conn net.Conn) {
	if t.cfg.AckWait <= 0 {
		return
	}
	errEnv := &transportpb.Envelope{Type: transportpb.MessageType_MESSAGE_TYPE_ERROR}
	if data, err := proto.Marshal(errEnv); err == nil {
		_ = writeLenPrefixed(conn, data)
	}
}

// generateEphemeralTLSConfig creates a self-signed Ed25519 certificate for this process lifetime.
func generateEphemeralTLSConfig(local transport.DeviceDescriptor) (tls.Certificate, *tls.Config, error) {
	pub, priv, err := ed25519.GenerateKey(cryptorand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	serial, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(1<<62))
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().Add(12 * time.Hour), // ephemeral lifetime
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"device." + local.DeviceID},
	}
	der, err := x509.CreateCertificate(cryptorand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	conf := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13, InsecureSkipVerify: true}
	return cert, conf, nil
}

// pkixName builds a minimal subject (avoid importing pkix just for CN string formatting).
func (t *Transport) serverTLSConfig() *tls.Config { return t.tlsConfig }
func (t *Transport) clientTLSConfig() *tls.Config { return t.tlsConfig }

// PinnedPublicKeys returns a snapshot copy of currently pinned peer public keys (DER bytes).
func (t *Transport) PinnedPublicKeys() map[string][]byte {
	out := map[string][]byte{}
	t.pinned.Range(func(key, value any) bool {
		id, ok1 := key.(string)
		b, ok2 := value.([]byte)
		if ok1 && ok2 {
			out[id] = append([]byte(nil), b...)
		}
		return true
	})
	return out
}

// writeLenPrefixed writes a uint32 length + data.
func writeLenPrefixed(w io.Writer, b []byte) error {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(b)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func readLenPrefixed(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	ln := binary.BigEndian.Uint32(hdr[:])
	if ln == 0 {
		return nil, nil
	}
	buf := make([]byte, ln)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

// NOTE: wireEnvelope replaced by protobuf Envelope (transportpb.Envelope)

// generateTLSConfigFromKey builds a TLS config given an existing Ed25519 private key.
func generateTLSConfigFromKey(local transport.DeviceDescriptor, priv ed25519.PrivateKey) (tls.Certificate, *tls.Config, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return tls.Certificate{}, nil, errors.New("lan: invalid ed25519 private key length")
	}
	pub := priv.Public().(ed25519.PublicKey)
	serial, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(1<<62))
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"device." + local.DeviceID},
	}
	der, err := x509.CreateCertificate(cryptorand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	conf := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13, InsecureSkipVerify: true}
	return cert, conf, nil
}

func init() { transport.Register("lan", factory{}) }

// advertAdapter wraps a generic AdvertisementHandle so tests can inject custom behavior.
type advertAdapter struct{ h transport.AdvertisementHandle }

func (a *advertAdapter) ServiceRecord() *zeroconf.ServiceRecord { return nil }
func (a *advertAdapter) Shutdown()                              { _ = a.h.Close() }
func (a *advertAdapter) Close()                                 { _ = a.h.Close() }

type zeroconfWrapper struct{ *zeroconf.Server }

func (z zeroconfWrapper) Close() { z.Server.Shutdown() }

// --- Protobuf conversion helpers ---
func toProtoBundle(b *sharing.SecretExportBundle) *transportpb.SecretExportBundle {
	if b == nil {
		return nil
	}
	pb := &transportpb.SecretExportBundle{Signature: b.Signature}
	pb.Payload = &transportpb.SecretExportPayload{
		Id:                     b.Payload.ID,
		Name:                   b.Payload.Name,
		EncryptedSecrets:       b.Payload.EncryptedSecrets,
		SecretsNonce:           b.Payload.SecretsNonce,
		SymmetricKeyBox:        b.Payload.SymmetricKeyBox,
		Timestamp:              b.Payload.Timestamp,
		ExpiresAt:              b.Payload.ExpiresAt,
		SenderDeviceName:       b.Payload.SenderInfo.DeviceName,
		SenderUserId:           b.Payload.SenderInfo.UserID,
		SenderPublicKey:        b.Payload.SenderInfo.PublicKey,
		SenderSigningPublicKey: b.Payload.SenderInfo.SigningPublicKey,
	}
	return pb
}

func fromProtoBundle(pb *transportpb.SecretExportBundle) *sharing.SecretExportBundle {
	if pb == nil || pb.Payload == nil {
		return nil
	}
	b := &sharing.SecretExportBundle{Signature: pb.Signature}
	p := pb.Payload
	b.Payload = sharing.SecretExportPayload{
		ID:               p.Id,
		Name:             p.Name,
		EncryptedSecrets: p.EncryptedSecrets,
		SecretsNonce:     p.SecretsNonce,
		SymmetricKeyBox:  p.SymmetricKeyBox,
		Timestamp:        p.Timestamp,
		ExpiresAt:        p.ExpiresAt,
		SenderInfo: sharing.SenderMetadata{
			DeviceName:       p.SenderDeviceName,
			UserID:           p.SenderUserId,
			PublicKey:        p.SenderPublicKey,
			SigningPublicKey: p.SenderSigningPublicKey,
		},
	}
	return b
}

// startAdvertisement publishes an mDNS service advertising this device.
func startAdvertisement(local *transport.DeviceDescriptor, addr net.Addr) (*zeroconf.Server, error) {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return nil, errors.New("lan: unexpected listener addr type")
	}
	port := tcpAddr.Port
	txt := []string{"device_id=" + local.DeviceID, "device_name=" + local.DeviceName}
	if len(local.Ed25519Pub) > 0 {
		txt = append(txt, "ed25519="+base64.StdEncoding.EncodeToString(local.Ed25519Pub))
	}
	srv, err := zeroconf.Register("vibes-"+shortID(local.DeviceID), "_vibes-pass._tcp", "local.", port, txt, nil)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
