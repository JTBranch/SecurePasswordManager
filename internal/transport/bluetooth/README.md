Bluetooth transport skeleton

This package implements a transport using BLE (GATT) to exchange encrypted SecretExportBundle payloads on supported platforms.

Platform support

- Linux: supported (BlueZ + go-ble/linux driver)
- macOS: supported (CoreBluetooth + go-ble/darwin driver)
- Windows: intentionally disabled in this branch — Bluetooth code is excluded from Windows builds.

Why Windows is disabled

- Windows exposes Bluetooth LE via WinRT/COM APIs rather than a POSIX-like interface; there is no widely supported, battle-tested pure-Go BLE driver for Windows that we can safely rely on today.
- To avoid pulling heavy or experimental WinRT bindings into the main binary, this project excludes the BLE implementation from Windows builds.

Path to add Windows support later

1. Native helper approach (recommended):

   - Implement a small native helper (C#/.NET or C++) that uses Windows WinRT Bluetooth APIs to perform scanning, connect, read, write, and subscribe operations.
   - Use a compact IPC protocol (stdin/stdout JSON or a named pipe) between the Go process and helper.
   - Add a Windows Go adapter that spawns/connects to the helper and implements the `Adapter` interface by translating adapter calls to helper requests.
   - This keeps the main app in pure Go while using platform-native APIs where appropriate.

2. Go WinRT bindings (advanced):
   - Use or create Go bindings to WinRT to call Gatt APIs directly from Go.
   - This removes the helper binary but is more complex and requires careful binding maintenance and cross-compilation support.

If you want, I can scaffold the native helper + Go adapter in `internal/transport/bluetooth/adapters/windows_helper/` and a Windows adapter that uses it.
