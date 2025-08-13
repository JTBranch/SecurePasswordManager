# Go Password Manager

## Overview
Go Password Manager is a simple, secure, and encrypted password manager designed for macOS and Windows. It allows users to store key-value secrets with built-in versioning and local storage of encrypted items. This application does not have any web or network functionality, ensuring that all data remains local and secure.

## Features
- **Secure Storage**: All secrets are stored in an encrypted format to protect sensitive information.
- **Versioning**: Keep track of changes to secrets with built-in versioning, allowing retrieval of previous versions.
- **Local Storage**: All data is stored locally on the user's machine, ensuring privacy and security.
- **Cross-Platform**: Compatible with both macOS and Windows.

## Project Structure
```
go-password-manager
├── cmd
│   └── main.go          # Entry point of the application
├── internal
│   ├── storage
│   │   └── storage.go   # Local storage implementation for encrypted items
│   ├── crypto
│   │   └── crypto.go    # Encryption and decryption functions
│   ├── secrets
│   │   └── secrets.go    # Management of key-value secrets
│   └── versioning
│       └── versioning.go # Versioning implementation for secrets
├── ui
│   ├── app.go           # User interface setup
│   └── components.go     # UI components definitions
├── go.mod               # Module definition
├── go.sum               # Module checksums
└── README.md            # Project documentation
```

## Installation
1. Clone the repository:
   ```
   git clone https://github.com/yourusername/go-password-manager.git
   ```
2. Navigate to the project directory:
   ```
   cd go-password-manager
   ```
3. Install dependencies:
   ```
   go mod tidy
   ```

## Usage
To run the application, execute the following command:
```
go run cmd/main.go
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for details.