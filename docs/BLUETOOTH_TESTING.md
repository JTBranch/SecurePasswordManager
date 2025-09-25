# PlatformAdapter Test Guide

This document explains the in-memory `PlatformAdapter` used for bluetooth integration tests and how to migrate away from the legacy `btmock` test adapter.

## Overview

- The `PlatformAdapter` is implemented at `internal/transport/bluetooth/platform_adapter.go` and provides an in-memory mailbox-per-device model suitable for headless integration tests.
- It supports the following helpers used across tests:
  - `NewPlatformAdapter()` – construct a new adapter instance
  - `CreateDevice(id)` – register a mailbox for `id`
  - `AdvertiseDevice(id)` / `StopAdvertise(id)` – mark a device as discoverable
  - `Scan(serviceUUID, limit)` – return advertised device IDs
  - `SetMTU(mtu)` – test-only: set MTU to force fragmentation
  - `FailNextWrites(n)` – test-only: simulate transient write failures for `n` writes

## Why this adapter

- It removes the need for platform BLE hardware during CI and local test runs.
- It models MTU fragmentation and transient failures so integration tests remain realistic.

## Testing helpers

- There's a small test helper in `tests/integration/bluetooth/helper.go` providing:
  - `SetupPairedPlatform(aID, bID)` – returns a `PlatformAdapter` and `DeviceDescriptor`s for two devices
  - `SetupAdvertisedDevice(adapter, id)` – creates and advertises a device and returns its descriptor

## Migration notes

- If your tests previously used the legacy `tests/integration/bluetooth/btmock` package, update them to use the helpers above and inject the `PlatformAdapter` into `transport.Dependencies{BluetoothAdapter: platform}` when building transports.
- The old `btmock` package has been removed from the repo.

## Example

Simple setup used in tests:

```go
platform := bt.NewPlatformAdapter()
platform.CreateDevice("sender")
platform.CreateDevice("receiver")
platform.AdvertiseDevice("receiver")
depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}
depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}

trA, _ := transport.Build(ctx, "bluetooth", nil, sender, depsA)
trB, _ := transport.Build(ctx, "bluetooth", nil, receiver, depsB)
// trA.Send(...), trB.Receive(...)
```

If you'd like a migration script or automated replacements across tests, ask and I can add one.
