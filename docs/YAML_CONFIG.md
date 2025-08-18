# YAML Configuration System

The Go Password Manager now supports environment-specific configuration using YAML files. This provides a more structured and maintainable way to manage configuration across different environments.

## Configuration Structure

Configuration files are located in the `configs/` directory:

```
configs/
├── default.yaml      # Base configuration for all environments
├── development.yaml  # Development-specific overrides
├── production.yaml   # Production-specific overrides
└── test.yaml        # Test environment overrides
```

## Configuration Schema

```yaml
application:
  name: "GoPasswordManager"
  version: "1.0.0"
  environment: "development"

ui:
  window:
    width: 1600
    height: 900
  theme: "default"

logging:
  debug: true
  level: "info"
  format: "json"

security:
  encryption:
    key_size: 32
    algorithm: "AES-256-GCM"

storage:
  secrets_file: "secrets.json"
  config_file: "app.config"

development:
  hot_reload: false
  auto_save: true

testing:
  timeout: "30s"
  data_dir: ""
  parallel: true
```

## Environment Configuration

### Setting the Environment

The environment is determined by the `GO_PASSWORD_MANAGER_ENV` environment variable:

```bash
# Development (default)
export GO_PASSWORD_MANAGER_ENV=development

# Production
export GO_PASSWORD_MANAGER_ENV=production

# Testing
export GO_PASSWORD_MANAGER_ENV=test
```

### Configuration Loading Order

1. **Default Configuration**: `configs/default.yaml` is loaded first
2. **Environment Configuration**: `configs/{environment}.yaml` is loaded and merged
3. **Environment Variables**: Environment variables override YAML values

### Environment Variable Overrides

You can override any configuration value using environment variables:

| Environment Variable | YAML Path | Example |
|---------------------|-----------|---------|
| `APP_NAME` | `application.name` | `GoPasswordManager` |
| `APP_VERSION` | `application.version` | `1.0.0` |
| `DEFAULT_WINDOW_WIDTH` | `ui.window.width` | `1600` |
| `DEFAULT_WINDOW_HEIGHT` | `ui.window.height` | `900` |
| `DEBUG_LOGGING` | `logging.debug` | `true` |
| `LOG_LEVEL` | `logging.level` | `debug` |
| `ENCRYPTION_KEY_SIZE` | `security.encryption.key_size` | `32` |
| `SECRETS_FILE_PATH` | `storage.secrets_file` | `secrets.json` |
| `CONFIG_FILE_PATH` | `storage.config_file` | `app.config` |
| `HOT_RELOAD` | `development.hot_reload` | `true` |
| `TEST_DATA_DIR` | `testing.data_dir` | `/tmp/test` |
| `E2E_TEST_TIMEOUT` | `testing.timeout` | `30s` |

## Usage in Code

### Loading Configuration

```go
import "go-password-manager/internal/envconfig"

// Load configuration
config, err := envconfig.Load()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Get global config instance
config := envconfig.Get()
```

### Environment Detection

```go
if config.IsDevelopment() {
    // Development-specific code
}

if config.IsProduction() {
    // Production-specific code
}

if config.IsTest() {
    // Test-specific code
}
```

### Accessing Configuration Values

```go
// Application settings
appName := config.Application.Name
version := config.Application.Version

// UI settings
width, height := config.GetWindowSize()

// Logging settings
debugEnabled := config.Logging.Debug
logLevel := config.Logging.Level

// File paths
secretsPath := config.GetSecretsFilePath()
configPath := config.GetConfigFilePath()

// Test settings
timeout := config.GetTestTimeout()
```

## Migration from Environment Variables

The system is designed to be backward compatible. Existing `.env` files and environment variables continue to work, with the YAML configuration system providing additional structure and organization.

### Migration Steps

1. **Create YAML configs**: Start by creating `configs/default.yaml` with your base configuration
2. **Environment-specific overrides**: Create environment-specific YAML files for values that differ between environments
3. **Gradual migration**: Move configuration from `.env` files to YAML files over time
4. **Remove `.env` dependency**: Once all configuration is in YAML, you can remove `.env` file dependencies

## Benefits

1. **Structure**: YAML provides better organization and readability
2. **Environment-specific**: Easy to manage different configurations per environment
3. **Type Safety**: Strong typing in Go structs
4. **Validation**: Built-in validation and error handling
5. **Documentation**: Self-documenting configuration structure
6. **Backward Compatibility**: Works alongside existing environment variable system

## Best Practices

1. **Keep secrets out of YAML**: Use environment variables for sensitive data
2. **Version control**: Commit YAML files to version control (except sensitive overrides)
3. **Environment-specific files**: Use `.env.local` or environment variables for machine-specific overrides
4. **Default values**: Always provide sensible defaults in `default.yaml`
5. **Testing**: Create dedicated test configurations for reliable testing
