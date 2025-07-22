package certstore

import (
	"crypto/x509"
	"fmt"
	"time"
)

// CertificateStore defines the interface for certificate store operations
type CertificateStore interface {
	// Name returns the name of the certificate store
	Name() string
	
	// IsSupported checks if this store is supported on the current platform
	IsSupported() bool
	
	// RequiresRoot returns true if root privileges are required
	RequiresRoot() bool
	
	// ListCertificates returns all certificates currently in the store
	ListCertificates() ([]*x509.Certificate, error)
	
	// AddCertificate adds a certificate to the store
	AddCertificate(cert *x509.Certificate) error
	
	// RemoveCertificate removes a certificate from the store
	RemoveCertificate(cert *x509.Certificate) error
	
	// Backup creates a backup of the current store state
	Backup(backupPath string) error
	
	// Restore restores the store from a backup
	Restore(backupPath string) error
	
	// Validate checks if the store is in a valid state
	Validate() error
}

// CertificateInfo contains metadata about a certificate
type CertificateInfo struct {
	Certificate   *x509.Certificate
	Subject       string
	Issuer        string
	SerialNumber  string
	NotBefore     time.Time
	NotAfter      time.Time
	Fingerprint   string
	IsCA          bool
	Source        string
}

// StoreType represents different types of certificate stores
type StoreType string

const (
	StoreTypeSystem      StoreType = "system"
	StoreTypeApplication StoreType = "application"
	StoreTypeCustom      StoreType = "custom"
)

// StoreFactory creates certificate store instances
type StoreFactory interface {
	CreateStore(storeType StoreType, target string, options map[string]string) (CertificateStore, error)
	SupportedStores() []string
}

// StoreManager manages multiple certificate stores
type StoreManager struct {
	stores   map[string]CertificateStore
	factory  StoreFactory
	verbose  bool
}

// NewStoreManager creates a new store manager
func NewStoreManager(factory StoreFactory, verbose bool) *StoreManager {
	return &StoreManager{
		stores:  make(map[string]CertificateStore),
		factory: factory,
		verbose: verbose,
	}
}

// AddStore adds a certificate store to the manager
func (sm *StoreManager) AddStore(name string, store CertificateStore) {
	sm.stores[name] = store
}

// GetStore retrieves a certificate store by name
func (sm *StoreManager) GetStore(name string) (CertificateStore, bool) {
	store, exists := sm.stores[name]
	return store, exists
}

// ListStores returns all managed stores
func (sm *StoreManager) ListStores() map[string]CertificateStore {
	return sm.stores
}

// CreateAndAddStore creates a new store and adds it to the manager
func (sm *StoreManager) CreateAndAddStore(name string, storeType StoreType, target string, options map[string]string) error {
	store, err := sm.factory.CreateStore(storeType, target, options)
	if err != nil {
		return fmt.Errorf("failed to create store %s: %w", name, err)
	}
	
	if !store.IsSupported() {
		return fmt.Errorf("store %s is not supported on this platform", name)
	}
	
	sm.AddStore(name, store)
	return nil
}

// ValidateAllStores validates all managed stores
func (sm *StoreManager) ValidateAllStores() error {
	for name, store := range sm.stores {
		if err := store.Validate(); err != nil {
			return fmt.Errorf("validation failed for store %s: %w", name, err)
		}
	}
	return nil
}

// BackupAllStores creates backups for all managed stores
func (sm *StoreManager) BackupAllStores(backupDir string) error {
	for name, store := range sm.stores {
		backupPath := fmt.Sprintf("%s/%s_backup_%d", backupDir, name, time.Now().Unix())
		if err := store.Backup(backupPath); err != nil {
			return fmt.Errorf("backup failed for store %s: %w", name, err)
		}
		if sm.verbose {
			fmt.Printf("Created backup for store %s at %s\n", name, backupPath)
		}
	}
	return nil
}
