# Trust Store Updater - Architecture Summary

## Original Prompt

> Implement a cross platform tool in Go which can update operating system and application trusts stores (ca bundles etc) with recommended new root certificates. It should target linux, macos and windows and use configuration to decide which target applicaiton stores or OS stores to update. Save this prompt as part of the architecture summary.

## Architecture Overview

The Trust Store Updater is a comprehensive cross-platform tool designed to manage certificate trust stores across different operating systems and applications. The architecture follows Go best practices with clear separation of concerns and platform-specific abstractions.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Interface                             │
│                     (Cobra Commands)                            │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                   Updater Service                               │
│              (Orchestration Layer)                             │
└─────┬─────────────────────────────────────────┬─────────────────┘
      │                                         │
┌─────▼─────────────────┐                ┌─────▼─────────────────┐
│  Certificate Fetcher  │                │   Store Manager      │
│  (HTTP, File, Dir)    │                │  (Multi-store ops)   │
└───────────────────────┘                └─────┬─────────────────┘
                                               │
                                    ┌─────────▼─────────┐
                                    │  Platform Factory │
                                    │ (Platform Router) │
                                    └─┬─────────┬───────┬─┘
                              ┌───────▼─┐ ┌─────▼─┐ ┌───▼───┐
                              │ Linux   │ │Darwin │ │Windows│
                              │Platform │ │Platform│ │Platform│
                              └─────────┘ └───────┘ └───────┘
```

### Core Components

#### 1. Command Line Interface (`internal/cmd/`)
- **Framework**: Cobra CLI framework
- **Configuration**: Viper for configuration management
- **Features**: Dry-run mode, verbose output, custom config files
- **Entry Point**: `cmd/trust-store-updater/main.go`

#### 2. Configuration System (`internal/config/`)
- **Format**: YAML-based configuration
- **Features**: 
  - Certificate source definitions (URL, file, directory)
  - Trust store targeting with platform filtering
  - Global settings (backup, validation, timeouts)
  - Automatic default configuration generation

#### 3. Certificate Management (`internal/cert/`)
- **Fetcher**: Multi-source certificate retrieval
  - HTTP/HTTPS endpoints with custom headers
  - Local file and directory scanning
  - PEM/DER format support
- **Validation**: Certificate validation and filtering
- **Utilities**: Fingerprinting, comparison, format conversion

#### 4. Certificate Store Abstraction (`internal/certstore/`)
- **Interface**: Common `CertificateStore` interface
- **Operations**: List, Add, Remove, Backup, Restore, Validate
- **Management**: Store manager for multi-store operations
- **Factory Pattern**: Platform-specific store creation

#### 5. Platform Implementations (`internal/platform/`)

##### Linux (`internal/platform/linux/`)
- **System Stores**:
  - `ca-certificates` (Debian/Ubuntu style)
  - `update-ca-trust` (RHEL/CentOS style)
- **Application Stores**:
  - Docker certificates
  - Java cacerts
  - Firefox certificate database
  - Chrome certificate handling

##### macOS (`internal/platform/darwin/`)
- **System Stores**:
  - System Keychain
  - Login Keychain
- **Application Stores**:
  - Docker certificates
  - Java cacerts
  - Firefox certificate database
  - Chrome certificate handling
  - Safari (uses system keychain)

##### Windows (`internal/platform/windows/`)
- **System Stores**:
  - Root certificate store
  - Intermediate CA store
  - Personal certificate store
  - Enterprise Trust store
- **Application Stores**:
  - Docker certificates
  - Java cacerts
  - Firefox certificate database
  - Chrome certificate handling
  - Microsoft Edge
  - IIS certificate store

#### 6. Update Orchestration (`internal/updater/`)
- **Service Layer**: Main update orchestration
- **Process Flow**:
  1. Configuration validation
  2. Platform-appropriate store initialization
  3. Optional backup creation
  4. Certificate fetching from all sources
  5. Store-by-store updates
  6. Post-update validation

### Key Design Patterns

#### 1. Interface Segregation
- `CertificateStore` interface defines common operations
- Platform-specific implementations provide concrete behavior
- Factory pattern creates appropriate store instances

#### 2. Dependency Injection
- Store manager accepts factory interface
- Updater service accepts configuration and dependencies
- Promotes testability and modularity

#### 3. Configuration-Driven Behavior
- YAML configuration controls:
  - Which certificate sources to use
  - Which trust stores to target
  - Platform-specific filtering
  - Operational settings (backup, validation, etc.)

#### 4. Error Handling Strategy
- Graceful degradation: Continue processing other stores on individual failures
- Comprehensive logging with configurable verbosity
- Backup mechanisms for recovery

### Security Considerations

#### 1. Privilege Management
- Automatic detection of required privileges per store type
- Graceful handling of insufficient permissions
- Clear documentation of privilege requirements

#### 2. Certificate Validation
- Basic certificate validation (expiry, CA status, basic constraints)
- Fingerprint-based duplicate detection
- Source verification for HTTPS fetches

#### 3. Backup and Recovery
- Automatic backup creation before modifications
- Store-specific backup strategies
- Manual restore capability

### Configuration Schema

```yaml
certificate_sources:
  - name: string          # Unique identifier
    type: url|file|directory
    source: string        # URL, file path, or directory path
    enabled: boolean
    verify_tls: boolean   # For URL sources
    headers: map         # For URL sources
    filters: []string    # File filters or certificate filters

trust_stores:
  - name: string          # Unique identifier
    type: system|application
    platform: []string   # [linux, darwin, windows]
    target: string       # Store-specific identifier
    enabled: boolean
    require_root: boolean
    options: map         # Store-specific options

settings:
  backup_enabled: boolean
  backup_directory: string
  log_level: string
  max_retries: integer
  timeout_seconds: integer
  validate_after: boolean
```

### Platform-Specific Store Identifiers

#### Linux
- **System**: `ca-certificates`, `update-ca-trust`
- **Application**: `docker`, `java-cacerts`, `firefox`, `chrome`

#### macOS
- **System**: `system-keychain`, `login-keychain`
- **Application**: `docker`, `java-cacerts`, `firefox`, `chrome`, `safari`

#### Windows
- **System**: `root`, `ca`, `my`, `trust`
- **Application**: `docker`, `java-cacerts`, `firefox`, `chrome`, `edge`, `iis`

### Implementation Status

#### Completed
- ✅ Project structure and module setup
- ✅ Configuration system with YAML support
- ✅ CLI interface with Cobra
- ✅ Certificate fetching from multiple sources
- ✅ Platform abstraction layer
- ✅ Store interface definitions
- ✅ Update orchestration logic
- ✅ Basic validation and error handling

#### In Progress / Future Work
- 🔄 Platform-specific store implementations (skeletons created)
- 🔄 Certificate validation enhancements
- 🔄 Comprehensive testing suite
- 🔄 Windows Certificate Store API integration
- 🔄 macOS Security framework integration
- 🔄 Linux package manager integration

### Dependencies

#### Core Dependencies
- `github.com/spf13/cobra`: CLI framework
- `github.com/spf13/viper`: Configuration management
- `gopkg.in/yaml.v3`: YAML parsing

#### Standard Library Usage
- `crypto/x509`: Certificate parsing and validation
- `net/http`: HTTP client for certificate fetching
- `os/exec`: System command execution
- `path/filepath`: Cross-platform path handling

### Build and Deployment

#### Build Commands
```bash
# Standard build
go build -o trust-store-updater ./cmd/trust-store-updater

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o trust-store-updater-linux ./cmd/trust-store-updater
GOOS=darwin GOARCH=amd64 go build -o trust-store-updater-darwin ./cmd/trust-store-updater
GOOS=windows GOARCH=amd64 go build -o trust-store-updater.exe ./cmd/trust-store-updater
```

#### VS Code Integration
- Build tasks configured in `.vscode/tasks.json`
- Copilot instructions in `.github/copilot-instructions.md`
- Go module support with `go.mod`

### Future Enhancements

1. **Enhanced Validation**: Certificate chain validation, CRL checking
2. **Rollback Mechanism**: Automatic rollback on validation failures
3. **Monitoring Integration**: Metrics and monitoring capabilities
4. **Plugin System**: Support for custom store implementations
5. **GUI Interface**: Optional graphical user interface
6. **Containerization**: Docker container support
7. **Package Managers**: Integration with system package managers

This architecture provides a solid foundation for cross-platform certificate trust store management while maintaining flexibility for future enhancements and platform-specific requirements.
