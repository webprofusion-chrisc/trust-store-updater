# Trust Store Updater

A cross-platform tool written in Go for updating operating system and application trust stores with new root certificates.

Note: this tool is a prototype and not yet suitable for production use. As Uncle Ben/Voltaire would say, with great power comes great responsibility (yours).

## Features

- **Cross-platform support**: Linux, macOS, and Windows
- **Configuration-driven**: Uses YAML configuration to specify which trust stores to update
- **Multiple certificate sources**: Fetch certificates from URLs, files, or directories
- **Backup and restore**: Automatic backup creation before updates
- **Dry-run mode**: Test updates without making changes
- **Comprehensive validation**: Certificate validation before installation
- **Flexible targeting**: Update system and application trust stores

## Architecture

The tool follows a modular architecture with clear separation of concerns:

```
cmd/
├── trust-store-updater/     # CLI application entry point
internal/
├── certstore/               # Certificate store interfaces and management
├── config/                  # Configuration handling
├── platform/                # Platform-specific implementations
│   ├── linux/              # Linux certificate store implementations
│   ├── darwin/             # macOS certificate store implementations
│   └── windows/            # Windows certificate store implementations
├── cert/                   # Certificate fetching and validation
├── updater/                # Main update orchestration logic
└── cmd/                    # CLI command definitions
```

## Supported Trust Stores

### Linux
- **System stores**: ca-certificates, update-ca-trust
- **Applications**: Docker, Java cacerts, Firefox, Chrome

### macOS
- **System stores**: System Keychain, Login Keychain
- **Applications**: Docker, Java cacerts, Firefox, Chrome, Safari

### Windows
- **System stores**: Root, CA, Personal, Enterprise Trust
- **Applications**: Docker, Java cacerts, Firefox, Chrome, Edge, IIS

## Installation

### Prerequisites
- Go 1.21 or later
- Platform-specific tools (varies by target store)

### Build from source
```bash
go mod download
go build -o trust-store-updater ./cmd/trust-store-updater
```

### Install dependencies
```bash
go mod tidy
```

## Usage

### Basic usage
```bash
# Update trust stores using default configuration
./trust-store-updater

# Use custom configuration file
./trust-store-updater --config /path/to/config.yaml

# Dry run to see what would be changed
./trust-store-updater --dry-run

# Verbose output
./trust-store-updater --verbose
```

### Configuration

The tool uses a YAML configuration file (`trust-store-config.yaml` by default). If the file doesn't exist, a default configuration will be created.

#### Example configuration:
```yaml
# Certificate sources - where to fetch new root certificates from
certificate_sources:
  - name: "mozilla-ca-bundle"
    type: "url"
    source: "https://curl.se/ca/cacert.pem"
    enabled: true
    verify_tls: true

  - name: "local-certificates"
    type: "directory"
    source: "./certificates"
    enabled: false
    filters:
      - "*.crt"
      - "*.pem"

# Trust stores - target stores to update with new certificates
trust_stores:
  - name: "system-ca-certificates"
    type: "system"
    platform: ["linux"]
    target: "ca-certificates"
    enabled: true
    require_root: true

  - name: "system-keychain"
    type: "system"
    platform: ["darwin"]
    target: "system-keychain"
    enabled: true
    require_root: true

  - name: "system-cert-store"
    type: "system"
    platform: ["windows"]
    target: "root"
    enabled: true
    require_root: true

# Global settings
settings:
  backup_enabled: true
  backup_directory: "./backups"
  log_level: "info"
  max_retries: 3
  timeout_seconds: 30
  validate_after: true
```

### Certificate Sources

The tool supports fetching certificates from multiple sources:

- **URL**: Fetch CA bundle from HTTP/HTTPS endpoints
- **File**: Load certificates from local PEM/DER files
- **Directory**: Scan directory for certificate files

### Trust Store Types

- **System stores**: Operating system certificate stores
- **Application stores**: Application-specific certificate stores

## Security Considerations

- **Root privileges**: Many system store operations require administrator/root privileges
- **Backup creation**: Always creates backups before making changes (configurable)
- **Certificate validation**: Validates certificates before installation
- **TLS verification**: Verifies TLS connections when fetching from URLs

## Development

### Project Structure
- Uses Go modules for dependency management
- Follows standard Go project layout
- Platform-specific code isolated in separate packages
- Interfaces used for abstraction and testability

### Key Dependencies
- `github.com/spf13/cobra`: CLI framework
- `github.com/spf13/viper`: Configuration management
- `gopkg.in/yaml.v3`: YAML parsing

### Building for Different Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o trust-store-updater-linux ./cmd/trust-store-updater

# macOS
GOOS=darwin GOARCH=amd64 go build -o trust-store-updater-darwin ./cmd/trust-store-updater

# Windows
GOOS=windows GOARCH=amd64 go build -o trust-store-updater.exe ./cmd/trust-store-updater
```

## Limitations

- Some platform-specific implementations are still in development
- Requires appropriate system permissions for trust store modifications
- Certificate validation is basic (no chain validation)
- No automatic rollback mechanism (manual restore from backup required)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Original Requirements

This tool was created to implement a cross-platform solution for updating operating system and application trust stores (CA bundles) with recommended new root certificates. It targets Linux, macOS, and Windows, using configuration to decide which target application stores or OS stores to update.

Key requirements addressed:
- Cross-platform compatibility (Linux, macOS, Windows)
- Configuration-driven trust store targeting
- Support for both OS and application trust stores
- Automated certificate fetching and validation
- Backup and restore capabilities
- Comprehensive error handling and logging
