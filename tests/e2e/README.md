# E2E Test Documentation

## Overview

This directory contains End-to-End (E2E) tests organized using the Page Object Model pattern for better maintainability and readability.

## Structure

```
tests/e2e/
├── setup/
│   └── test_setup.go      # Test environment setup and configuration
├── pages/
│   └── main_page.go       # Page Object Model for main application page
├── service_layer_crud_test.go    # Service layer CRUD operations tests
└── crud_operations_test.go       # UI-based CRUD operations tests (placeholder)
```

## Test Categories

### 1. Service Layer Tests (`service_layer_crud_test.go`)

These tests focus on the service layer functionality without UI interaction:

- **TestSecretCRUDOperationsServiceLayer**: Complete CRUD workflow

  - Create a secret on first load
  - Verify persistence between app restarts
  - Test versioning functionality
  - Test secret deletion

- **TestSecretCRUDOperationsErrorHandling**: Error scenarios
  - Error handling for non-existent secrets
  - Multiple secrets management

### 2. UI Tests (`crud_operations_test.go`)

Placeholder for future UI automation tests using the Page Object Model:

- **TestSecretCRUDOperations**: UI-based CRUD operations
- **TestSecretVisibilityToggle**: UI interaction testing

## Page Object Model

The `pages/main_page.go` file contains the `MainPageObject` class that provides:

- UI element interaction methods
- Business logic abstractions
- Consistent test interface

Example usage:

```go
mainPage := pages.NewMainPageObject(t, window, secretsService)
mainPage.LoadPage()
mainPage.ClickCreateSecretButton()
mainPage.FillCreateSecretModal("SecretName", "SecretValue")
mainPage.SubmitCreateSecretModal()
```

## Test Environment

The `setup/test_setup.go` provides:

- Isolated test environments using temporary directories
- Environment variable management
- Test data cleanup
- Application lifecycle management

## Running Tests

```bash
# Run all E2E tests
make e2eTest

# Run specific test pattern
go test ./tests/e2e -run TestSecretCRUDOperationsServiceLayer -v

# Run with verbose output
go test ./tests/e2e -v
```

## Features Tested

✅ **Working Tests:**

- Secret creation and persistence
- App restart simulation with data persistence
- Multiple secrets management
- Secret encryption/decryption
- Error handling for invalid operations

🔄 **In Development:**

- UI automation with Fyne framework
- Complete versioning workflow
- Advanced error scenarios

## Test Isolation

Each test suite:

- Uses isolated temporary directories
- Has independent encryption keys
- Cleans up automatically after completion
- Can share data directories for cross-test scenarios

## Future Enhancements

1. **UI Automation**: Complete Fyne UI interaction implementation
2. **Performance Tests**: Load testing with large numbers of secrets
3. **Cross-Platform Tests**: Test on different operating systems
4. **Integration Tests**: Database and file system integration
5. **Security Tests**: Encryption and key management validation
