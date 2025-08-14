# Go Password Manager

A secure, cross-platform password manager built in Go with a modern Fyne UI. All secrets are encrypted and stored locally, plans for third party secret conversion and perhaps a browser extension in the future. Will always be free, password security should be for everyone

## Features

- Atomic design UI structure (pages, molecules, atoms)
- Secure secrets storage and encryption
- Configurable window size and persistent UI state
- Unit and E2E test support via Makefile
- Easy extensibility and modular codebase

## Getting Started

### Prerequisites

- Go 1.20+
- [Fyne](https://fyne.io/) (UI library)

### Installation

Clone the repo:

```sh
git clone https://github.com/JTBranch/SecurePasswordManager.git
cd SecurePasswordManager
```

Install dependencies:

```sh
go mod tidy
```

### Environment Variables

You can control where secrets are stored using the `GO_PASSWORD_MANAGER_ENV` environment variable:

- **Development/Testing (default):**
  - `secrets.json` is stored in the project root and tracked by git (for easy inspection).
- **Production:**
  - Set `GO_PASSWORD_MANAGER_ENV=prod` before running the app.
  - `secrets.json` will be stored in your OS user config directory:
    - **macOS/Linux:** `~/Library/Application Support/GoPasswordManager/secrets.json` (macOS) or `~/.config/GoPasswordManager/secrets.json` (Linux)
    - **Windows:** `%APPDATA%\GoPasswordManager\secrets.json`
  - This file is never tracked by git and is protected by OS file permissions.

#### Example: Running in Production

```sh
export GO_PASSWORD_MANAGER_ENV=prod
go run cmd/main.go
```

#### Example: Running in Development

```sh
go run cmd/main.go
```

#### Development: Watch Mode

For rapid development, you can run the app in watch mode so it automatically reloads when you change files. This is typically done using a tool like [`air`](https://github.com/air-verse/air):

1. Install air (if not already):

   ```sh
   go install github.com/air-verse/air@latest
   ```

2. Add Go's bin directory to your PATH (optional, for easier usage):

   ```sh
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

3. Run in watch mode:

   ```sh
   air
   ```

   Or if you didn't add to PATH:

   ```sh
   $(go env GOPATH)/bin/air
   ```

This will watch your Go files and restart the app on changes. By default, this runs in development mode (with secrets.json in the repo).

### Testing

#### Unit Tests

Run all unit tests:

```sh
make unitTest
```

#### E2E Tests

Run all end-to-end tests:

```sh
make e2eTest
```

## Project Structure

```
go-password-manager/
  cmd/                # Main entrypoint
  internal/           # Core logic (config, crypto, service, domain)
  ui/
    pages/            # Top-level UI pages
    molecules/        # UI components (header, modals, etc)
    atoms/            # Smallest UI elements
    e2e/              # E2E UI tests
  tests/
    e2e/              # E2E workflow tests
  Makefile            # Test commands
```

## Contributing

We welcome contributions! Please:

- Fork the repo and create a feature branch
- Follow the atomic design and modular code guidelines
- Write unit and E2E tests for new features
- Submit a pull request with a clear description

For major changes, please open an issue first to discuss what you’d like to change.

## License

MIT
