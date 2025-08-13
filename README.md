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

### Running the App

```sh
go run cmd/main.go
```

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
