# 🔐 Go Password Manager - Easy Installation

A secure, cross-platform password manager with a modern GUI interface.

## 📥 Quick Install (Recommended)

### 🍎 **macOS** (Mac users)

```bash
curl -fsSL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/install-macos.sh | bash
```

Notes: the macOS installer now creates a native `.app` bundle in `~/Applications/PasswordManager/PasswordManager.app` so the system shows the correct Dock icon. The installer also sets the application environment to `production` when launched via the provided shortcuts.

### 🪟 **Windows** (Windows users)

1. Download: [install-windows.bat](https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/install-windows.bat)
2. Right-click the file → "Run as administrator"
3. Follow the installation prompts

### 🐧 **Linux** (Linux users)

```bash
curl -fsSL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/install-linux.sh | bash
```

## 🚀 Running the App

After installation, you can run the password manager:

### 🍎 **macOS**

- Double-click "Password Manager.command" on your Desktop
- Or run from Terminal: `~/Applications/PasswordManager/password-manager`

Note: On macOS the installer now creates `PasswordManager.app`. To open the packaged app run:

```bash
open ~/Applications/PasswordManager/PasswordManager.app
```

### 🪟 **Windows**

- Double-click "Password Manager" shortcut on your Desktop
- Or navigate to: `%USERPROFILE%\AppData\Local\PasswordManager\password-manager.exe`

### 🐧 **Linux**

- Find "Go Password Manager" in your applications menu
- Or run from Terminal: `password-manager`

## 📋 Manual Download

If you prefer to download manually:

1. Go to [Releases](https://github.com/JTBranch/SecurePasswordManager/releases/latest)
2. Download the file for your platform:
   - **macOS Apple Silicon**: `go-password-manager-macos-arm64`
   - **macOS Intel**: `go-password-manager-macos-amd64`
   - **Windows**: `password-manager-windows-amd64.exe`
   - **Linux**: `password-manager-linux-amd64`
3. Make it executable and run

## ⚡ Features

- **🔒 Strong Encryption**: AES-256 encryption for all stored passwords
- **📱 Modern GUI**: Clean, intuitive Fyne-based interface
- **🔄 Version Control**: Track password history and changes
- **🖥️ Cross-Platform**: Works on macOS, Windows, and Linux
- **📁 Portable**: Self-contained binary with local storage

## 🛡️ Security

- All passwords are encrypted using AES-256-GCM
- Master password never stored in plain text
- Secrets stored locally (no cloud sync)
- Open source for transparency

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/JTBranch/SecurePasswordManager/issues)
- **Documentation**: [Wiki](https://github.com/JTBranch/SecurePasswordManager/wiki)
- **Releases**: [All Versions](https://github.com/JTBranch/SecurePasswordManager/releases)

## 🏗️ For Developers

```bash
# Clone and build from source
git clone https://github.com/JTBranch/SecurePasswordManager.git
cd SecurePasswordManager
make build
```

Smoke test after installing (example):

```bash
./scripts/smoke-test-installed.sh ~/Applications/PasswordManager/PasswordManager.app/Contents/MacOS/password-manager
```

---

**Made with ❤️ and Go** | [GitHub](https://github.com/JTBranch/SecurePasswordManager) | [License](LICENSE)
