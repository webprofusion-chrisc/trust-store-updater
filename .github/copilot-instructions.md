<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Trust Store Updater - Copilot Instructions

This is a cross-platform tool written in Go for updating operating system and application trust stores with new root certificates.

## Project Architecture
- **Target Platforms**: Linux, macOS, Windows
- **Configuration-driven**: Uses YAML configuration to specify which trust stores to update
- **Modular design**: Platform-specific implementations with common interfaces
- **Certificate management**: Handles downloading, validation, and installation of root certificates

## Key Components
- `cmd/`: CLI application entry point using Cobra
- `internal/config/`: Configuration management
- `internal/certstore/`: Certificate store interfaces and implementations
- `internal/platform/`: Platform-specific implementations (linux, darwin, windows)
- `internal/cert/`: Certificate handling and validation
- `pkg/`: Public APIs and utilities

## Development Guidelines
- Use interfaces for platform abstraction
- Implement comprehensive error handling
- Include proper logging throughout
- Follow Go best practices for error handling and package organization
- Use dependency injection for testability
- Implement proper certificate validation before installation

## Trust Store Targets
- **Linux**: ca-certificates, update-ca-trust, application-specific stores
- **macOS**: Keychain, system trust settings
- **Windows**: Certificate Store API, specific application stores
- **Applications**: Browser stores, Docker, Java keystores, etc.
