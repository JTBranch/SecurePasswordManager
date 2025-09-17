# Secret Sharing Protocol (Current Version: v2.1)

## Export Flow (v2.1)

2. Construct two AEAD AAD domains:
   - Key Box AAD: `bundleID || 0x00 || sender_signing_public_key`
   - Secrets AAD: `bundleID || 0x01 || sender_signing_public_key` (domain‑separated by tag 0x01)
3. Encrypt secrets JSON with `symKey` + Secrets AAD using AES-GCM -> `EncryptedSecrets` + `SecretsNonce`.
4. Create key box: generate ephemeral X25519 key pair; derive shared secret; AES-GCM encrypt `symKey` with Key Box AAD -> `SymmetricKeyBox`.
5. Populate payload fields and sign with sender Ed25519 private key.
6. Return `SecretExportBundle`.

## Import Flow (v2.1)

2. Reconstruct Key Box AAD: `bundleID || 0x00 || sender_signing_public_key`; unwrap `SymmetricKeyBox` using recipient X25519 private key.
3. Reconstruct Secrets AAD: `bundleID || 0x01 || sender_signing_public_key`; decrypt `EncryptedSecrets` with `symKey` and `SecretsNonce`.
4. Replay defense: refuse bundle if `bundleID` previously processed within replay TTL window (default 1h) before persisting.
5. Deserialize and persist.

## Security Considerations (v2.1)

- **Replay**: Pluggable replay defense. Default: in-memory TTL cache prevents re-import of identical `bundleID` within configured window (default 1 hour). Optional persistent file-backed store (`FileReplayStore`) survives restarts. Choose implementation based on deployment profile (see Replay Store Configuration below).
- **Associated Data**: Dual-domain AEAD:
  - Key Box AAD (`0x00` tag) binds the transported symmetric key to bundle ID + signing key.
  - Secrets AAD (`0x01` tag) separately binds secrets ciphertext to the same context, preventing cross-protocol or structure substitution.

## Future Improvements (post v2.1)

| Persistent Replay Cache | External storage (file/db) for seen bundle IDs | Survive restarts; stronger replay defense |
| Secrets Metadata AAD | Optionally extend secrets AAD to include select metadata fields | Additional authenticated context |
| Secure Wipe | Wipe symmetric key after bundle assembly (already implemented) | Reduce key residency time in memory |
Updated: 2025-09-17 (v2.1: dual AAD + replay protection + persistent store option)

# Secret Sharing Protocol (Current Version)

## Overview

The sharing subsystem packages exported secrets plus a wrapped symmetric key into a single bundle. The symmetric key used to encrypt the JSON-serialized secrets list is randomly generated per export (32 bytes, AES-256) and transported via a compact _Key Box_ structure.

## Key Box Format

```
version (1 byte) | ephemeral_public_key (32 bytes) | nonce (12 bytes) | ciphertext (len >= 16)
```

- `version`: currently `0x02`.
- `ephemeral_public_key`: X25519 public key derived from a freshly generated 32-byte scalar.
- `nonce`: GCM nonce (12 bytes) generated with `crypto/rand`.
- `ciphertext`: AES-GCM ciphertext of the 32‑byte symmetric secret-encryption key (v2 uses AEAD AAD binding described below).

## Algorithms

- Asymmetric key agreement: X25519 (Curve25519).
- Symmetric encryption (secrets + key box AEAD): AES-256-GCM.
- Key derivation: Not used in new design for sharing (HKDF path removed). Each export uses a freshly generated symmetric key.
- Signatures: Ed25519 over JSON marshalled `SecretExportPayload` (excluding the signature field).

## Export Flow

1. Generate random 32-byte symmetric key (`symKey`) via `crypto/rand`.
2. Encrypt secrets JSON with `symKey` using AES-GCM -> `EncryptedSecrets` + `SecretsNonce`.
3. Create key box: generate ephemeral X25519 key pair; derive shared secret; build AAD (bundleID || 0x00 || sender_signing_public_key); AES-GCM encrypt `symKey` with AAD -> `SymmetricKeyBox`.
4. Populate payload fields and sign with sender Ed25519 private key.
5. Return `SecretExportBundle`.

## Import Flow

1. Validate expiry (UNIX seconds) and signature.
2. Reconstruct AAD and unwrap `SymmetricKeyBox` using recipient X25519 private key obtaining `symKey` (AEAD authenticity includes AAD binding).
3. Decrypt `EncryptedSecrets` using `symKey` and `SecretsNonce`.
4. Deserialize and persist.

## Zeroization / Memory Hygiene

Implemented (v2):

- Ephemeral private scalar zeroed after shared secret derivation (wrap path).
- Shared secret zeroed after use in both wrap and unwrap.
- Symmetric key slice zeroed after wrapping and after unwrapping (callers wipe once material used).

Potential additional hardening (future):

- Secure buffer abstraction with page locking (platform-dependent).
- Constant-time version dispatch if multiple active versions coexist.

## Security Considerations

- **Freshness / Forward Secrecy**: Each export uses a new ephemeral X25519 key and random symmetric key; compromise of a single export does not aid decryption of others.
- **Integrity**: Provided by AES-GCM tags (with AAD in v2) and Ed25519 signature. Tampering with any payload field (or key box bytes) causes decryption or signature verification failure.
- **Replay**: (Superseded by v2.1 section) Use pluggable replay store. Memory mode for ephemeral/local; file mode for persistent / production. Consider DB-backed store for multi-node deployments.
- **Associated Data**: Key box v2 binds bundle ID + sender signing public key via AEAD AAD (bundleID || 0x00 || signingPub). Secrets ciphertext currently has no external AAD; may add selected metadata later.
- **Side Channels**: Operations rely on well-reviewed X25519/AES-GCM implementations. Additional constant-time auditing could be performed for custom code paths (minimal at present).
- **Entropy**: All randomness from `crypto/rand`. No deterministic derivations in new path beyond curve operations.

## Backward Compatibility

The legacy fields `EncryptedSymmetricKey`, `KeyNonce`, and `EphemeralPublicKey` have been removed. v1 key boxes (version 0x01 without AAD) are now rejected; bundles must be re-exported under v2.

## Error Handling

- Unwrap validates minimum size ( >= 1 + 32 + 12 + 1 ) and version. Truncation or tampering yields errors; tests cover these cases.
- Distinct error messages intentionally minimal to avoid oracle style leakage (generic derivation / malformed messages).

## Future Improvements

| Area                 | Proposal                                                                | Rationale                                    |
| -------------------- | ----------------------------------------------------------------------- | -------------------------------------------- |
| Secrets AAD          | Add AAD (bundle ID + signing key) to secrets encryption                 | Binds metadata to secrets ciphertext         |
| Key Rotation         | Add explicit KDF salt & info fields (future v3)                         | Extensibility & domain separation            |
| Replay Defense       | Maintain seen bundle IDs until expiry                                   | Prevent re-import loops                      |
| Replay Store Scaling | Support database or distributed cache backends                          | Coordinate across multi-instance deployments |
| Secure Wipe          | Wipe symmetric key after bundle assembly                                | Reduce key residency time in memory          |
| Compression          | Optional pre-encryption compression (careful of CRIME/BREACH analogues) | Reduce payload size                          |
| Metadata Auth        | Include selected sender metadata fields as AAD for secrets encryption   | Stronger binding of metadata to ciphertext   |

## Testing Summary

Implemented tests for:

- Round-trip (wrap/unwrap) correctness (v2 with AAD binding).
- Wrong recipient key failure.
- Truncation and tamper failures.
- Unsupported version failure.

Integration tests validate full export/import path with updated schema.

---

Updated: 2025-09-17 (v2 with AAD)

## Replay Store Configuration

Two implementations are available:

1. In-Memory (`InMemoryReplayStore`)

- Volatile; cleared on process restart.
- Lowest latency, zero I/O.
- Ideal for: local development, short-lived CLI sessions, environments where replay risk across restarts is minimal.

2. File-Backed (`FileReplayStore`)

- JSON map persisted atomically via temp file rename.
- Survives restarts; still single-process oriented.
- Ideal for: single-node production, desktop app distribution, environments needing continuity after crash/restart.

Selection Strategy:

- Development / CI: use memory store for simplicity.
- Single-instance Production: use file store to harden against restarts.
- Multi-instance / Horizontal Scaling: implement a future `ReplayStore` backed by a shared DB or cache (e.g., BoltDB, SQLite WAL, Redis) to ensure global replay protection.

Operational Considerations:

- TTL Tuning: Shorter TTL lowers memory/disk usage; longer increases protection window. Default 1h balances accidental immediate replays vs. unbounded growth.
- Persistence File Path: Place in OS-specific app data directory; restrict permissions (0600) since bundle IDs may leak usage patterns.
- Rotation / Size Control: Future enhancement: max entries threshold triggering prune before flush; optional periodic flush goroutine.

Example (switch to file store):

```
importService.UseFileReplayStore("replay_store.json")
importService.SetReplayTTL(3600) // 1h
```

ReplayConfig (preferred simple integration):

```
cfg := sharing.ReplayConfig{Mode: sharing.ReplayModeFile, FilePath: "replay_store.json", TTL: time.Hour, MaxEntries: 10000, FlushInterval: 5 * time.Minute}
importService, err := sharing.NewImportServiceWithReplayConfig(cryptoProvider, deviceKeyProvider, secretsProvider, cfg)
if err != nil { /* handle */ }
```

Security Note:

- Replay entries are opaque identifiers; avoid embedding sensitive metadata in bundle IDs.
- For stronger guarantees across multiple hosts, do not rely on file store—design a shared authoritative store.
