# Sharing Transport Layer (Phase 1 Skeleton)

Goal: Pluggable, fault-tolerant mechanisms to move an exported `SecretExportBundle` between devices without modifying core sharing/crypto logic.

## Design Principles

1. Transport Agnostic Core: Export / Import remain unchanged; transports only carry the bundle.
2. Pluggability: Each transport registers a factory by ID (e.g. `lan`, `qr`, `email`).
3. Unified Interface: `BundleTransport` exposes `Send`, optional `Receive`, `Close`.
4. Idempotent / Replay-Safe: Bundle-level replay protection already enforced; transports may retry safely.
5. Minimal Dependencies: Standard library first; external libs only when essential.
6. Dependency Injection + Mockery: Interfaces are small & mock-friendly.

## Package Layout

```
internal/transport/
  transport.go        # interfaces + registry
  device.go           # device descriptors + registries
  errors.go
  lan/                # LAN skeleton (Phase 1)
    lan.go
    messages.go       # future handshake / framing definitions
```

Future additions:

```
internal/transport/qr/
internal/transport/email/
```

## Device Registry

Stores metadata (DeviceID, UserID, public keys, last-known address). Implementations:

- InMemoryRegistry (dev/test)
- FileRegistry (persistent desktop use)
  Later: multi-process (DB) or discovery service.

## LAN Transport (Current Implementation)

Pivot: We use standard library TLS 1.3 with self‑signed (or persisted) Ed25519 certificates instead of building a custom handshake first. Above TLS we now send a compact protobuf envelope instead of ad‑hoc JSON.

Key points:

1. Listener binds to `:0` by default (ephemeral port) and (optionally) advertises an mDNS service (`_gopass-pass._tcp`) when `discovery=true` in config.
2. Persistent identity: If a crypto service can supply / generate an Ed25519 key pair and the descriptor lacks one, a self‑signed cert is created from it; else an ephemeral cert (12–24h) is generated.
3. Protobuf framing: A length‑prefixed `Envelope` message is written: `type=DATA`, `sender_id`, `bundle` (serialized `SecretExportBundle`). ACK and ERROR are reserved envelope types.
4. Receiver pushes bundle to an internal channel and (if `AckWait>0`) responds with an `Envelope{ type=ACK }`.
5. Sender waits up to `AckWait` (default ~750ms) for ACK; timeout triggers retry with exponential backoff + jitter.

Discovery (Zeroconf / mDNS):

- When enabled, the transport publishes TXT keys: `device_id`, `device_name`, and optional `ed25519` (base64 public key).
- A lightweight browser (`lan/discovery.go`) runs asynchronous browsing and maintains an in‑memory cache of discovered `DeviceDescriptor`s (including resolved IP:port).
- The browser exposes `List()` for UI polling; stop via `Stop()` to release resources.

Security Notes:

- TLS 1.3 provides channel confidentiality/integrity; current trust is TOFU (no cert pinning yet). Future: compare presented Ed25519 against registry / TXT record for binding.
- Protobuf envelope constrains message surface (enum typed) vs free‑form JSON.

Planned Enhancements:

- Mutual device auth (certificate / Ed25519 binding) & pinning.
- ERROR envelope usage for richer negative acknowledgements.
- Optional encrypted discovery metadata (selectively omit public key in TXT until trust established).

## Future Transports

| Transport | Notes                                                                                                  |
| --------- | ------------------------------------------------------------------------------------------------------ |
| QR        | Chunked base64 / alphanumeric segments, sequence numbers, checksum, optional short-lived signing code. |
| Email     | ASCII armored bundle (header metadata + base64 payload, detached signature).                           |
| File      | Same as email w/o delivery medium.                                                                     |

## Fault Tolerance & Retries

Implemented (Phase 1.2):

- Protobuf envelope (DATA, ACK, ERROR placeholder) over length‑prefixed TLS stream.
- Exponential backoff retry with jitter (80–120% randomization per step) starting 50ms doubling per attempt.
- Optional ACK wait ensures positive delivery; missing or wrong envelope type triggers retry.
- Idempotent semantics rely on core replay protection.
- Discovery advertisement + passive device enumeration (optional, off by default unless configured).

Future:

- Distinguish between network vs application errors for adaptive retry.
- Passive health metrics (success ratio, average RTT) to inform UI hints.
- Negative ACK / error codes for richer sender feedback.

## Testing Strategy

- Unit: Factory registration, registry persistence, LAN build, stub send receipt.
- Integration (later): Loopback LAN end-to-end with real handshake once implemented.
- QR & Email: Deterministic encoding/decoding round trips; chunk ordering robustness.

## Mockery Targets (examples)

```
mockery --name BundleTransport --dir internal/transport --output internal/testHelpers/mocks
mockery --name DeviceRegistry --dir internal/transport --output internal/testHelpers/mocks
```

## Security Considerations (Upcoming)

- Mutual authentication of devices (Ed25519 signatures over handshake transcript).
- Ephemeral ECDH per session for forward secrecy.
- AEAD channel binding (bundleID not required at transport layer; prevents cross-channel replay).
- Optional rate limiting or proof-of-possession token for unsolicited inbound connections.

## Phase Roadmap

| Phase | Scope                                                 |
| ----- | ----------------------------------------------------- |
| 1     | Interfaces, registry, LAN TLS skeleton, docs (done)   |
| 2     | Persistent identity + retries + ACK (done)            |
| 3     | Protobuf framing + jitter + zeroconf discovery (done) |
| 4     | QR encoding (export only) + import scanner API        |
| 5     | Email/File armored format + signature verification    |
| 6     | Mutual device auth / pinning + metrics/log hooks      |
