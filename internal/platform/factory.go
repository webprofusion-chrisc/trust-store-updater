package platform

import (
	"fmt"
	"runtime"

	"github.com/trust-store-updater/internal/certstore"
	"github.com/trust-store-updater/internal/platform/darwin"
	"github.com/trust-store-updater/internal/platform/linux"
	"github.com/trust-store-updater/internal/platform/windows"
)

// Factory creates platform-specific certificate stores
type Factory struct {
	verbose bool
}

// NewFactory creates a new platform factory
func NewFactory(verbose bool) *Factory {
	return &Factory{verbose: verbose}
}

// CreateStore creates a certificate store based on the current platform
func (f *Factory) CreateStore(storeType certstore.StoreType, target string, options map[string]string) (certstore.CertificateStore, error) {
	switch runtime.GOOS {
	case "linux":
		return f.createLinuxStore(storeType, target, options)
	case "darwin":
		return f.createDarwinStore(storeType, target, options)
	case "windows":
		return f.createWindowsStore(storeType, target, options)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// SupportedStores returns a list of supported stores for the current platform
func (f *Factory) SupportedStores() []string {
	switch runtime.GOOS {
	case "linux":
		return linux.SupportedStores()
	case "darwin":
		return darwin.SupportedStores()
	case "windows":
		return windows.SupportedStores()
	default:
		return []string{}
	}
}

func (f *Factory) createLinuxStore(storeType certstore.StoreType, target string, options map[string]string) (certstore.CertificateStore, error) {
	switch storeType {
	case certstore.StoreTypeSystem:
		return linux.NewSystemStore(target, options, f.verbose)
	case certstore.StoreTypeApplication:
		return linux.NewApplicationStore(target, options, f.verbose)
	default:
		return nil, fmt.Errorf("unsupported store type for Linux: %s", storeType)
	}
}

func (f *Factory) createDarwinStore(storeType certstore.StoreType, target string, options map[string]string) (certstore.CertificateStore, error) {
	switch storeType {
	case certstore.StoreTypeSystem:
		return darwin.NewSystemStore(target, options, f.verbose)
	case certstore.StoreTypeApplication:
		return darwin.NewApplicationStore(target, options, f.verbose)
	default:
		return nil, fmt.Errorf("unsupported store type for macOS: %s", storeType)
	}
}

func (f *Factory) createWindowsStore(storeType certstore.StoreType, target string, options map[string]string) (certstore.CertificateStore, error) {
	switch storeType {
	case certstore.StoreTypeSystem:
		return windows.NewSystemStore(target, options, f.verbose)
	case certstore.StoreTypeApplication:
		return windows.NewApplicationStore(target, options, f.verbose)
	default:
		return nil, fmt.Errorf("unsupported store type for Windows: %s", storeType)
	}
}

// GetCurrentPlatform returns the current platform name
func GetCurrentPlatform() string {
	return runtime.GOOS
}

// IsPlatformSupported checks if a platform is supported
func IsPlatformSupported(platforms []string) bool {
	currentPlatform := GetCurrentPlatform()
	for _, platform := range platforms {
		if platform == currentPlatform {
			return true
		}
	}
	return false
}
