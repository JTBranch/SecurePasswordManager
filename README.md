# Go Password Manager

A secure, cross-platform password manager built in Go with a modern Fyne UI. All secrets are encrypted and stored locally with full version history. Features include secret editing, version management, and a clean atomic UI design. Plans for third party secret conversion and perhaps a browser extension in the future. Will always be free, password security should be for everyone.

## AI-Assisted Development

This project serves as an **experiment in AI-assisted software development**, exploring how modern AI tools can help build production-quality applications with robust automation, comprehensive testing, and high code quality standards. All major architectural decisions, code design patterns, and development direction have been made and supervised by the human developer, with AI assistance used as a productivity tool for implementation.

The development process leverages AI assistance for:

- **Architecture Implementation**: Translating architectural decisions into clean, modular patterns (atomic UI design, service layers)
- **Code Quality**: Automated formatting, linting, static analysis, and comprehensive test coverage
- **CI/CD Pipeline**: Automated builds, testing, coverage enforcement, and deployment workflows
- **Best Practices**: Implementing Go conventions, security patterns, and maintainable code structures

The goal is to demonstrate that AI-assisted development, under proper human guidance and supervision, can produce enterprise-grade software while maintaining code quality, security, and maintainability standards typically expected in production environments.

## Features

- **Secure Local Storage**: AES encrypted secrets stored locally
- **Version History**: Full version tracking with ability to view previous secret values
- **Secret Management**: Create, edit, view, and delete secrets with ease
- **Atomic UI Design**: Clean component structure (pages, molecules, atoms)
- **Development Tools**: Hot reload, comprehensive testing, easy build system
- **Cross-Platform**: Runs on macOS, Linux, and Windows
- **Environment-Aware**: Separate development and production storage locations

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

### Quick Start

The easiest way to get started is using the Makefile:

```sh
# Run in development mode
make dev

# Run with hot reload (install air first with `make install-deps`)
make dev-watch

# Build and run in production mode
make build
make run
```

### Available Commands

| Command             | Description                                |
| ------------------- | ------------------------------------------ |
| `make dev`          | Run in development mode with debug logging |
| `make dev-watch`    | Run with hot reload (requires air)         |
| `make build`        | Build the application binary               |
| `make run`          | Run the built binary in production mode    |
| `make unitTest`     | Run unit tests                             |
| `make e2eTest`      | Run end-to-end tests                       |
| `make clean`        | Remove build artifacts                     |
| `make install-deps` | Install development dependencies (air)     |

### Environment Variables

The application supports different environments controlled by the `GO_PASSWORD_MANAGER_ENV` variable:

- **Development Mode (default):**

  - `secrets.json` stored in project root
  - Debug logging enabled
  - Suitable for development and testing

- **Production Mode:**
  - Set `GO_PASSWORD_MANAGER_ENV=prod`
  - `secrets.json` stored in OS user config directory:
    - **macOS:** `~/Library/Application Support/GoPasswordManager/secrets.json`
    - **Linux:** `~/.config/GoPasswordManager/secrets.json`
    - **Windows:** `%APPDATA%\GoPasswordManager\secrets.json`
  - Minimal logging
  - Secure file permissions

### Data Structure

Secrets are stored with full version history in a nested JSON structure:

```json
{
  "secrets": [
    {
      "secretName": "example",
      "type": "key_value",
      "currentVersion": 2,
      "versions": [
        {
          "secretValueEnc": "encrypted_value_v1",
          "version": 1,
          "updatedAt": "2025-08-14T10:29:00+01:00"
        },
        {
          "secretValueEnc": "encrypted_value_v2",
          "version": 2,
          "updatedAt": "2025-08-14T10:30:00+01:00"
        }
      ]
    }
  ]
}
```

This structure allows you to:

- Track all changes to secrets over time
- View previous versions of any secret
- Maintain audit trail of when secrets were modified

### Testing

```sh
# Run unit tests
make unitTest

# Run end-to-end tests
make e2eTest

# Run all tests
make unitTest && make e2eTest
```

## Application Features

### Secret Management

- **Create Secrets**: Add new encrypted secrets with automatic versioning
- **Edit Secrets**: Modify existing secrets while preserving version history
- **View Secrets**: Reveal secret values with click-to-show functionality
- **Delete Secrets**: Remove secrets entirely (including all versions)

### Version History

- **Track Changes**: Every edit creates a new version with timestamp
- **View History**: Browse previous versions of any secret
- **Restore Values**: Copy previous versions for restoration
- **Audit Trail**: Complete history of when secrets were modified

### User Interface

- **Clean Design**: Atomic UI components for consistent experience
- **Responsive Layout**: Adapts to different window sizes
- **Intuitive Navigation**: Easy-to-use buttons and forms
- **Security First**: Values hidden by default with reveal functionality

## Project Structure

```
go-password-manager/
├── cmd/                    # Application entrypoint
│   └── main.go            # Main application file
├── internal/              # Core application logic
│   ├── config/            # Configuration management
│   ├── crypto/            # Encryption/decryption
│   ├── domain/            # Data models and types
│   ├── logger/            # Logging utilities
│   ├── service/           # Business logic services
│   ├── storage/           # File storage handling
│   └── versioning/        # Version management
├── ui/                    # User interface components
│   ├── atoms/             # Smallest UI components
│   ├── molecules/         # Composite UI components
│   ├── pages/             # Full page layouts
│   └── e2e/               # UI end-to-end tests
├── tests/                 # Test suites
│   └── e2e/               # End-to-end workflow tests
├── Makefile              # Build and development commands
├── .air.toml             # Hot reload configuration
└── README.md             # This file
```

### Architecture Principles

- **Atomic Design**: UI components organized in atoms → molecules → pages hierarchy
- **Clean Architecture**: Clear separation between domain, service, and UI layers
- **Security First**: All secrets encrypted at rest with AES encryption
- **Version Control**: Complete audit trail of all secret modifications
- **Environment Aware**: Separate development and production configurations

## Contributing

We welcome contributions! Please:

- Fork the repo and create a feature branch
- Follow the atomic design and modular code guidelines
- Write unit and E2E tests for new features
- Submit a pull request with a clear description

For major changes, please open an issue first to discuss what you’d like to change.

## License

MIT
