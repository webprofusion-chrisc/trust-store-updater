# Getting Started with Trust Store Updater

## Prerequisites Installation

### Installing Go

Since Go isn't currently installed on your system, you'll need to install it first:

#### Windows (your current platform)
1. **Download Go**: Visit https://golang.org/dl/ and download the Windows installer
2. **Install**: Run the MSI installer and follow the prompts
3. **Verify**: Open a new PowerShell window and run:
   ```powershell
   go version
   ```

#### Alternative: Using Chocolatey (if you have it)
```powershell
choco install golang
```

#### Alternative: Using winget
```powershell
winget install GoLang.Go
```

## Building the Project

Once Go is installed:

1. **Navigate to the project directory**:
   ```powershell
   cd "C:\Work\GIT\misc\trust-store-updater"
   ```

2. **Download dependencies**:
   ```powershell
   go mod download
   go mod tidy
   ```

3. **Build the application**:
   ```powershell
   go build -o trust-store-updater.exe ./cmd/trust-store-updater
   ```

4. **Or use the VS Code task**: Press `Ctrl+Shift+P`, type "Tasks: Run Task", and select "Build Trust Store Updater"

## Running the Application

### First Run
```powershell
# This will create a default configuration file
./trust-store-updater.exe --help
```

### Basic Usage
```powershell
# Dry run to see what would happen
./trust-store-updater.exe --dry-run --verbose

# Actual update (requires appropriate permissions)
./trust-store-updater.exe --verbose
```

### Custom Configuration
```powershell
# Use a custom config file
./trust-store-updater.exe --config my-config.yaml --dry-run
```

## Development

### Running Tests (when implemented)
```powershell
go test ./...
```

### Cross-Platform Builds
```powershell
# For Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o trust-store-updater-linux ./cmd/trust-store-updater

# For macOS
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o trust-store-updater-darwin ./cmd/trust-store-updater

# Reset environment
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

## Project Status

✅ **Completed**:
- Project structure and architecture
- Configuration system
- CLI interface
- Certificate fetching framework
- Platform abstraction layer
- Update orchestration logic

🔄 **In Development**:
- Platform-specific implementations (skeleton code created)
- Full certificate validation
- Comprehensive testing

## Next Steps

1. **Install Go** (see instructions above)
2. **Build and test** the basic functionality
3. **Implement platform-specific store handlers** as needed
4. **Add comprehensive tests**
5. **Deploy and configure** for your specific use case

## Troubleshooting

### Common Issues

1. **"go: command not found"**
   - Go is not installed or not in PATH
   - Restart PowerShell after installation

2. **Permission errors when updating system stores**
   - Run as Administrator for system-level changes
   - Some operations require elevated privileges

3. **Module download issues**
   - Check internet connectivity
   - Verify Go proxy settings: `go env GOPROXY`

### Getting Help

- Check the README.md for detailed usage instructions
- Review ARCHITECTURE.md for technical details
- Examine the configuration file comments for options
