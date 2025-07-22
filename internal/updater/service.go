package updater

import (
	"crypto/x509"
	"fmt"
	"os"
	"runtime"

	"github.com/trust-store-updater/internal/cert"
	"github.com/trust-store-updater/internal/certstore"
	"github.com/trust-store-updater/internal/config"
	"github.com/trust-store-updater/internal/platform"
)

// Service handles the certificate trust store update process
type Service struct {
	config      *config.Config
	storeManager *certstore.StoreManager
	fetcher     *cert.Fetcher
	verbose     bool
	dryRun      bool
}

// New creates a new updater service
func New(cfg *config.Config, verbose, dryRun bool) *Service {
	factory := platform.NewFactory(verbose)
	storeManager := certstore.NewStoreManager(factory, verbose)
	fetcher := cert.NewFetcher(cfg.Settings.TimeoutSeconds, verbose)

	return &Service{
		config:       cfg,
		storeManager: storeManager,
		fetcher:      fetcher,
		verbose:      verbose,
		dryRun:       dryRun,
	}
}

// UpdateTrustStores performs the trust store update process
func (s *Service) UpdateTrustStores() error {
	if s.verbose {
		fmt.Printf("Starting trust store update process (dry-run: %v)\n", s.dryRun)
		fmt.Printf("Platform: %s\n", runtime.GOOS)
	}

	// Validate configuration
	if err := config.ValidateConfig(s.config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize trust stores
	if err := s.initializeTrustStores(); err != nil {
		return fmt.Errorf("failed to initialize trust stores: %w", err)
	}

	// Create backup if enabled
	if s.config.Settings.BackupEnabled && !s.dryRun {
		if err := s.createBackups(); err != nil {
			return fmt.Errorf("failed to create backups: %w", err)
		}
	}

	// Fetch certificates from all sources
	allCerts, err := s.fetchAllCertificates()
	if err != nil {
		return fmt.Errorf("failed to fetch certificates: %w", err)
	}

	if s.verbose {
		fmt.Printf("Fetched %d certificates from all sources\n", len(allCerts))
	}

	// Update each trust store
	for name, store := range s.storeManager.ListStores() {
		if err := s.updateStore(name, store, allCerts); err != nil {
			fmt.Printf("Warning: Failed to update store %s: %v\n", name, err)
			continue
		}
	}

	// Validate stores after update
	if s.config.Settings.ValidateAfter && !s.dryRun {
		if err := s.storeManager.ValidateAllStores(); err != nil {
			return fmt.Errorf("post-update validation failed: %w", err)
		}
	}

	if s.verbose {
		fmt.Println("Trust store update completed successfully")
	}

	return nil
}

// initializeTrustStores sets up all configured trust stores
func (s *Service) initializeTrustStores() error {
	currentPlatform := platform.GetCurrentPlatform()

	for _, storeConfig := range s.config.TrustStores {
		if !storeConfig.Enabled {
			if s.verbose {
				fmt.Printf("Skipping disabled store: %s\n", storeConfig.Name)
			}
			continue
		}

		// Check if this store is supported on current platform
		if !platform.IsPlatformSupported(storeConfig.Platform) {
			if s.verbose {
				fmt.Printf("Skipping store %s: not supported on platform %s\n", storeConfig.Name, currentPlatform)
			}
			continue
		}

		// Check root privileges if required
		if storeConfig.RequireRoot && os.Geteuid() != 0 {
			fmt.Printf("Warning: Store %s requires root privileges, skipping\n", storeConfig.Name)
			continue
		}

		// Create store
		storeType := certstore.StoreType(storeConfig.Type)
		err := s.storeManager.CreateAndAddStore(storeConfig.Name, storeType, storeConfig.Target, storeConfig.Options)
		if err != nil {
			fmt.Printf("Warning: Failed to create store %s: %v\n", storeConfig.Name, err)
			continue
		}

		if s.verbose {
			fmt.Printf("Initialized store: %s (%s)\n", storeConfig.Name, storeConfig.Target)
		}
	}

	return nil
}

// createBackups creates backups of all stores
func (s *Service) createBackups() error {
	if s.verbose {
		fmt.Printf("Creating backups in directory: %s\n", s.config.Settings.BackupDirectory)
	}

	return s.storeManager.BackupAllStores(s.config.Settings.BackupDirectory)
}

// fetchAllCertificates fetches certificates from all configured sources
func (s *Service) fetchAllCertificates() (map[string][]*Certificate, error) {
	allCerts := make(map[string][]*Certificate)

	for _, source := range s.config.CertificateSources {
		if !source.Enabled {
			if s.verbose {
				fmt.Printf("Skipping disabled source: %s\n", source.Name)
			}
			continue
		}

		certs, err := s.fetchFromSource(source)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch from source %s: %v\n", source.Name, err)
			continue
		}

		if s.verbose {
			fmt.Printf("Fetched %d certificates from source: %s\n", len(certs), source.Name)
		}

		allCerts[source.Name] = certs
	}

	return allCerts, nil
}

// fetchFromSource fetches certificates from a single source
func (s *Service) fetchFromSource(source config.CertificateSource) ([]*Certificate, error) {
	var rawCerts []*x509.Certificate
	var err error

	switch source.Type {
	case "url":
		rawCerts, err = s.fetcher.FetchFromURL(source.Source, source.Headers, source.VerifyTLS)
	case "file":
		rawCerts, err = s.fetcher.FetchFromFile(source.Source)
	case "directory":
		rawCerts, err = s.fetcher.FetchFromDirectory(source.Source, source.Filters)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}

	if err != nil {
		return nil, err
	}

	// Filter certificates
	filteredCerts := cert.FilterCertificates(rawCerts, source.Filters)

	// Convert to our certificate type and validate
	var validCerts []*Certificate
	for _, rawCert := range filteredCerts {
		if err := s.fetcher.ValidateCertificate(rawCert); err != nil {
			if s.verbose {
				fmt.Printf("Warning: Certificate validation failed for %s: %v\n", rawCert.Subject.CommonName, err)
			}
			continue
		}

		certInfo := &Certificate{
			X509Cert: rawCert,
			Source:   source.Name,
			Info:     cert.GetCertificateInfo(rawCert),
		}
		validCerts = append(validCerts, certInfo)
	}

	return validCerts, nil
}

// updateStore updates a single trust store with certificates
func (s *Service) updateStore(name string, store certstore.CertificateStore, allCerts map[string][]*Certificate) error {
	if s.verbose {
		fmt.Printf("Updating store: %s\n", name)
	}

	if s.dryRun {
		fmt.Printf("DRY RUN: Would update store %s with certificates\n", name)
		return nil
	}

	// Get current certificates in store
	currentCerts, err := store.ListCertificates()
	if err != nil {
		return fmt.Errorf("failed to list current certificates: %w", err)
	}

	// Collect all new certificates
	var newCerts []*Certificate
	for _, sourceCerts := range allCerts {
		newCerts = append(newCerts, sourceCerts...)
	}

	// Determine which certificates to add
	toAdd := s.findCertificatesToAdd(currentCerts, newCerts)

	if s.verbose {
		fmt.Printf("Adding %d new certificates to store %s\n", len(toAdd), name)
	}

	// Add new certificates
	for _, certToAdd := range toAdd {
		if err := store.AddCertificate(certToAdd.X509Cert); err != nil {
			fmt.Printf("Warning: Failed to add certificate %s to store %s: %v\n", 
				certToAdd.X509Cert.Subject.CommonName, name, err)
		} else if s.verbose {
			fmt.Printf("Added certificate: %s\n", certToAdd.X509Cert.Subject.CommonName)
		}
	}

	return nil
}

// findCertificatesToAdd determines which certificates need to be added
func (s *Service) findCertificatesToAdd(currentCerts []*x509.Certificate, newCerts []*Certificate) []*Certificate {
	var toAdd []*Certificate

	for _, newCert := range newCerts {
		found := false
		for _, currentCert := range currentCerts {
			if cert.CompareCertificates(newCert.X509Cert, currentCert) {
				found = true
				break
			}
		}

		if !found {
			toAdd = append(toAdd, newCert)
		}
	}

	return toAdd
}

// Certificate represents a certificate with metadata
type Certificate struct {
	X509Cert *x509.Certificate
	Source   string
	Info     map[string]interface{}
}
