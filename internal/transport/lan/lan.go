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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go-password-manager/internal/logger"
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
	cfg          Config
	local        transport.DeviceDescriptor
	deps         transport.Dependencies
	ln           net.Listener
	closed       chan struct{}
	tlsCert      tls.Certificate
	tlsConfig    *tls.Config
	inbound      chan *transport.InboundBundle
	advertiser   interface{ Close() }
	pinned       sync.Map // deviceID -> DER SubjectPublicKeyInfo
	fallbackPath string   // path to fallback instance file
}

// Addr returns the listen address of the transport (for tests / diagnostics).
func (t *Transport) Addr() string {
	if t.ln == nil {
		return ""
	}
	return t.ln.Addr().String()
}

func (t *Transport) ID() string { return "lan" }

// Discover returns currently pinned peer devices (TOFU) as a lightweight
// discovery mechanism. Future enhancement: active zeroconf browse.
const (
	txtDeviceID   = "device_id="
	txtDeviceName = "device_name="
	serviceType   = "_gopass-pass._tcp"
	serviceDomain = "local."
)

func (t *Transport) Discover(ctx context.Context, limit int) ([]transport.DeviceDescriptor, error) {
	logger.Debug("lan.discover: start")
	var fresh []transport.DeviceDescriptor
	if t.cfg.Discovery {
		fresh = t.activeBrowseCollect(ctx, limit)
	} else {
		logger.Debug("lan.discover: discovery disabled in config")
	}
	dedup := map[string]transport.DeviceDescriptor{}
	for _, d := range fresh {
		if d.DeviceID != t.local.DeviceID {
			dedup[d.DeviceID] = d
		}
	}
	t.pinned.Range(func(key, value any) bool {
		if id, ok := key.(string); ok && id != t.local.DeviceID {
			if _, exists := dedup[id]; !exists {
				dedup[id] = transport.DeviceDescriptor{DeviceID: id, DeviceName: id}
			}
		}
		return true
	})
	results := make([]transport.DeviceDescriptor, 0, len(dedup))
	for _, d := range dedup {
		results = append(results, d)
	}
	if len(results) == 0 { // fallback only when no mDNS results
		fb := t.readFallbackPeers()
		for _, d := range fb {
			if d.DeviceID != t.local.DeviceID {
				results = append(results, d)
			}
		}
	}
	logger.Debug("lan.discover: aggregated devices=" + itoa(len(results)))
	if limit > 0 && len(results) > limit {
		return results[:limit], nil
	}
	return results, nil
}

func (t *Transport) activeBrowseCollect(ctx context.Context, limit int) []transport.DeviceDescriptor {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		logger.Debug("lan.discover: resolver error: " + err.Error())
		return nil
	}
	entries := make(chan *zeroconf.ServiceEntry)
	browseCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	logger.Debug("lan.discover: browsing _gopass-pass._tcp for up to 2000ms")
	if err := resolver.Browse(browseCtx, serviceType, serviceDomain, entries); err != nil {
		logger.Debug("lan.discover: browse error: " + err.Error())
		cancel()
		return nil
	}
	devices := map[string]transport.DeviceDescriptor{}
Loop:
	for {
		select {
		case e, ok := <-entries:
			if !ok {
				logger.Debug("lan.discover: entries channel closed")
				break Loop
			}
			if e == nil {
				continue
			}
			logger.Debug("lan.discover: entry name=" + e.Instance + " port=" + itoa(e.Port))
			devID, devName := parseTXT(e.Text)
			logger.Debug("lan.discover: parsed devID=" + devID + " devName=" + devName)
			if devID == "" || devID == t.local.DeviceID {
				continue
			}
			if devName == "" {
				devName = devID
			}
			if _, exists := devices[devID]; !exists {
				dd := transport.DeviceDescriptor{DeviceID: devID, DeviceName: devName}
				if len(e.AddrIPv4) > 0 {
					dd.LastAddr = e.AddrIPv4[0].String() + ":" + itoa(e.Port)
				} else if len(e.AddrIPv6) > 0 {
					dd.LastAddr = "[" + e.AddrIPv6[0].String() + "]:" + itoa(e.Port)
				}
				devices[devID] = dd
				logger.Debug("lan.discover: added device id=" + devID + " addr=" + dd.LastAddr)
				if _, seen := t.pinned.Load(devID); !seen {
					t.pinned.Store(devID, []byte("placeholder"))
				}
				if limit > 0 && len(devices) >= limit {
					cancel()
					break Loop
				}
			}
		case <-browseCtx.Done():
			logger.Debug("lan.discover: browse context done")
			break Loop
		}
	}
	cancel()
	res := make([]transport.DeviceDescriptor, 0, len(devices))
	for _, d := range devices {
		res = append(res, d)
	}
	logger.Debug("lan.discover: collected devices=" + itoa(len(res)))
	return res
}

// --- Fallback file-based discovery (development aid) ---
func (t *Transport) writeFallbackFile(addr string) {
	if addr == "" {
		return
	}
	dir := os.TempDir()
	path := filepath.Join(dir, "gopass-pass-instance-"+t.local.DeviceID)
	data := t.local.DeviceID + "|" + t.local.DeviceName + "|" + addr
	if err := os.WriteFile(path, []byte(data), 0o600); err == nil {
		logger.Debug("lan.fallback: wrote instance file=" + path)
		t.fallbackPath = path
	} else {
		logger.Debug("lan.fallback: write error: " + err.Error())
	}
}

func (t *Transport) readFallbackPeers() []transport.DeviceDescriptor {
	dir := os.TempDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	// Deduplicate by device name picking newest file; prune stale (>20s)
	const maxAge = 20 * time.Second
	now := time.Now()
	type candidate struct {
		d   transport.DeviceDescriptor
		mod time.Time
	}
	byName := map[string]candidate{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "gopass-pass-instance-") {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		age := now.Sub(info.ModTime())
		if age > maxAge { // prune stale file
			_ = os.Remove(filepath.Join(dir, name))
			continue
		}
		path := filepath.Join(dir, name)
		b, rerr := os.ReadFile(path)
		if rerr != nil || len(b) == 0 {
			continue
		}
		parts := strings.Split(string(b), "|")
		if len(parts) != 3 {
			continue
		}
		if parts[0] == t.local.DeviceID {
			continue
		} // skip self
		d := transport.DeviceDescriptor{DeviceID: parts[0], DeviceName: parts[1], LastAddr: parts[2]}
		if cur, ok := byName[d.DeviceName]; ok {
			if info.ModTime().After(cur.mod) {
				byName[d.DeviceName] = candidate{d: d, mod: info.ModTime()}
			}
		} else {
			byName[d.DeviceName] = candidate{d: d, mod: info.ModTime()}
		}
	}
	peers := make([]transport.DeviceDescriptor, 0, len(byName))
	for _, c := range byName {
		peers = append(peers, c.d)
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].DeviceName < peers[j].DeviceName })
	logger.Debug("lan.fallback: loaded peers=" + itoa(len(peers)))
	return peers
}

// parseTXT extracts device id and name from zeroconf TXT records.
func parseTXT(txts []string) (devID, devName string) {
	for _, txt := range txts {
		if strings.HasPrefix(txt, txtDeviceID) {
			devID = strings.TrimPrefix(txt, txtDeviceID)
			continue
		}
		if strings.HasPrefix(txt, txtDeviceName) {
			devName = strings.TrimPrefix(txt, txtDeviceName)
		}
	}
	return
}

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
		if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
			ip := tcpAddr.IP
			if ip == nil || ip.IsUnspecified() {
				ip = pickInterfaceIPv4()
			}
			addrStr := ip.String() + ":" + itoa(tcpAddr.Port)
			if deps.Advertisement != nil {
				logger.Debug("lan.advertise: starting injected advertisement addr=" + addrStr)
				if handle, err2 := deps.Advertisement.Start(&t.local, ip.String(), tcpAddr.Port, nil); err2 == nil {
					logger.Debug("lan.advertise: injected advertisement started")
					t.advertiser = &advertAdapter{h: handle}
				} else {
					logger.Debug("lan.advertise: injected advertisement error: " + err2.Error())
				}
			} else if adv, err2 := startAdvertisementWithIP(&t.local, ip, tcpAddr.Port); err2 == nil {
				logger.Debug("lan.advertise: zeroconf register success device=" + t.local.DeviceID + " addr=" + addrStr)
				t.advertiser = zeroconfWrapper{Server: adv}
			} else {
				logger.Debug("lan.advertise: zeroconf register error: " + err2.Error())
			}
			// fallback file uses concrete ip:port
			t.writeFallbackFile(addrStr)
		}
	}
	go t.acceptLoop()
	return t, nil
}

// pickInterfaceIPv4 chooses first non-loopback UP interface IPv4, fallback loopback.
func pickInterfaceIPv4() net.IP {
	ifs, err := net.Interfaces()
	if err != nil {
		return net.ParseIP("127.0.0.1")
	}
	for _, iface := range ifs {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.To4() != nil {
				return ipNet.IP.To4()
			}
		}
	}
	return net.ParseIP("127.0.0.1")
}

// startAdvertisementWithIP registers with explicit interface IP.
func startAdvertisementWithIP(local *transport.DeviceDescriptor, ip net.IP, port int) (*zeroconf.Server, error) {
	txt := []string{"device_id=" + local.DeviceID, "device_name=" + local.DeviceName}
	if len(local.Ed25519Pub) > 0 {
		txt = append(txt, "ed25519="+base64.StdEncoding.EncodeToString(local.Ed25519Pub))
	}
	// Use default interface selection by passing nil for ifaces; zeroconf binds to all.
	return zeroconf.Register("vibes-"+shortID(local.DeviceID), serviceType, serviceDomain, port, txt, nil)
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
// startAdvertisement removed; replaced by startAdvertisementWithIP simplifying interface selection.

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
