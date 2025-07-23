package linux

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/webprofusion/trust-store-updater/internal/certstore"
)

// SystemStore implements certificate store operations for Linux system stores
type SystemStore struct {
	target  string
	options map[string]string
	verbose bool
}

// NewSystemStore creates a new Linux system certificate store
func NewSystemStore(target string, options map[string]string, verbose bool) (certstore.CertificateStore, error) {
	store := &SystemStore{
		target:  target,
		options: options,
		verbose: verbose,
	}

	// Validate target
	if !isValidSystemTarget(target) {
		return nil, fmt.Errorf("unsupported system store target: %s", target)
	}

	return store, nil
}

// Name returns the name of the certificate store
func (s *SystemStore) Name() string {
	return fmt.Sprintf("linux-system-%s", s.target)
}

// IsSupported checks if this store is supported on the current platform
func (s *SystemStore) IsSupported() bool {
	switch s.target {
	case "ca-certificates":
		return s.hasCaCertificates()
	case "update-ca-trust":
		return s.hasUpdateCaTrust()
	default:
		return false
	}
}

// RequiresRoot returns true if root privileges are required
func (s *SystemStore) RequiresRoot() bool {
	return true
}

// ListCertificates returns all certificates currently in the store
func (s *SystemStore) ListCertificates() ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	switch s.target {
	case "ca-certificates":
		return s.listCaCertificates()
	case "update-ca-trust":
		return s.listUpdateCaTrustCertificates()
	default:
		return certs, fmt.Errorf("unsupported target: %s", s.target)
	}
}

// AddCertificate adds a certificate to the store
func (s *SystemStore) AddCertificate(cert *x509.Certificate) error {
	switch s.target {
	case "ca-certificates":
		return s.addCaCertificate(cert)
	case "update-ca-trust":
		return s.addUpdateCaTrustCertificate(cert)
	default:
		return fmt.Errorf("unsupported target: %s", s.target)
	}
}

// RemoveCertificate removes a certificate from the store
func (s *SystemStore) RemoveCertificate(cert *x509.Certificate) error {
	switch s.target {
	case "ca-certificates":
		return s.removeCaCertificate(cert)
	case "update-ca-trust":
		return s.removeUpdateCaTrustCertificate(cert)
	default:
		return fmt.Errorf("unsupported target: %s", s.target)
	}
}

// Backup creates a backup of the current store state
func (s *SystemStore) Backup(backupPath string) error {
	switch s.target {
	case "ca-certificates":
		return s.backupCaCertificates(backupPath)
	case "update-ca-trust":
		return s.backupUpdateCaTrust(backupPath)
	default:
		return fmt.Errorf("unsupported target: %s", s.target)
	}
}

// Restore restores the store from a backup
func (s *SystemStore) Restore(backupPath string) error {
	switch s.target {
	case "ca-certificates":
		return s.restoreCaCertificates(backupPath)
	case "update-ca-trust":
		return s.restoreUpdateCaTrust(backupPath)
	default:
		return fmt.Errorf("unsupported target: %s", s.target)
	}
}

// Validate checks if the store is in a valid state
func (s *SystemStore) Validate() error {
	if !s.IsSupported() {
		return fmt.Errorf("store is not supported on this system")
	}

	// Check if we have required permissions
	if s.RequiresRoot() && os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required for system store operations")
	}

	return nil
}

// Helper methods

func isValidSystemTarget(target string) bool {
	validTargets := []string{"ca-certificates", "update-ca-trust"}
	for _, valid := range validTargets {
		if target == valid {
			return true
		}
	}
	return false
}

func (s *SystemStore) hasCaCertificates() bool {
	_, err := exec.LookPath("update-ca-certificates")
	return err == nil
}

func (s *SystemStore) hasUpdateCaTrust() bool {
	_, err := exec.LookPath("update-ca-trust")
	return err == nil
}

// listCaCertificatesFromDir lists all certificates from the specified directory (for testability)
func listCaCertificatesFromDir(certDir string) ([]*x509.Certificate, error) {
	files, err := os.ReadDir(certDir)
	if err != nil {
		certstore.LogErrorf("Failed to read cert dir: %v", err)
		return nil, fmt.Errorf("failed to read cert dir: %w", err)
	}

	var certs []*x509.Certificate
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !(strings.HasSuffix(name, ".crt") || strings.HasSuffix(name, ".pem")) {
			continue
		}
		path := filepath.Join(certDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			certstore.LogWarnf("Skipping unreadable file: %s (%v)", path, err)
			continue // skip unreadable files
		}
		// Parse all PEM blocks in the file
		rest := data
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err == nil {
					certs = append(certs, cert)
				} else {
					certstore.LogWarnf("Failed to parse certificate in %s: %v", path, err)
				}
			}
		}
	}
	certstore.LogInfof("Found %d certificates in %s", len(certs), certDir)
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in %s", certDir)
	}
	return certs, nil
}

func (s *SystemStore) listCaCertificates() ([]*x509.Certificate, error) {
	return listCaCertificatesFromDir("/etc/ssl/certs/")
}

func (s *SystemStore) listUpdateCaTrustCertificates() ([]*x509.Certificate, error) {
	// List all .pem/.crt files in /etc/pki/ca-trust/source/anchors/
	certDir := "/etc/pki/ca-trust/source/anchors/"
	files, err := os.ReadDir(certDir)
	if err != nil {
		certstore.LogErrorf("Failed to read anchors dir: %v", err)
		return nil, fmt.Errorf("failed to read anchors dir: %w", err)
	}

	var certs []*x509.Certificate
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !(strings.HasSuffix(name, ".crt") || strings.HasSuffix(name, ".pem")) {
			continue
		}
		path := filepath.Join(certDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			certstore.LogWarnf("Skipping unreadable file: %s (%v)", path, err)
			continue // skip unreadable files
		}
		// Parse all PEM blocks in the file
		rest := data
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err == nil {
					certs = append(certs, cert)
				} else {
					certstore.LogWarnf("Failed to parse certificate in %s: %v", path, err)
				}
			}
		}
	}
	certstore.LogInfof("Found %d certificates in %s", len(certs), certDir)
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in %s", certDir)
	}
	return certs, nil
}

func (s *SystemStore) addCaCertificate(cert *x509.Certificate) error {
	// Add certificate to /usr/local/share/ca-certificates/
	certDir := "/usr/local/share/ca-certificates/"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate a filename based on certificate subject
	filename := generateCertFilename(cert) + ".crt"
	certPath := filepath.Join(certDir, filename)

	// Write certificate to file
	if err := writeCertificateToFile(cert, certPath); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Update ca-certificates
	cmd := exec.Command("update-ca-certificates")
	if s.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update ca-certificates: %w", err)
	}

	return nil
}

func (s *SystemStore) addUpdateCaTrustCertificate(cert *x509.Certificate) error {
	// Add certificate to /etc/pki/ca-trust/source/anchors/
	certDir := "/etc/pki/ca-trust/source/anchors/"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate a filename based on certificate subject
	filename := generateCertFilename(cert) + ".crt"
	certPath := filepath.Join(certDir, filename)

	// Write certificate to file
	if err := writeCertificateToFile(cert, certPath); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Update ca-trust
	cmd := exec.Command("update-ca-trust", "extract")
	if s.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update ca-trust: %w", err)
	}

	return nil
}

func (s *SystemStore) removeCaCertificate(cert *x509.Certificate) error {
	// Remove certificate from /usr/local/share/ca-certificates/
	filename := generateCertFilename(cert) + ".crt"
	certPath := filepath.Join("/usr/local/share/ca-certificates/", filename)

	if err := os.Remove(certPath); err != nil {
		return fmt.Errorf("failed to remove certificate: %w", err)
	}

	// Update ca-certificates
	cmd := exec.Command("update-ca-certificates")
	if s.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func (s *SystemStore) removeUpdateCaTrustCertificate(cert *x509.Certificate) error {
	// Remove certificate from /etc/pki/ca-trust/source/anchors/
	filename := generateCertFilename(cert) + ".crt"
	certPath := filepath.Join("/etc/pki/ca-trust/source/anchors/", filename)

	if err := os.Remove(certPath); err != nil {
		return fmt.Errorf("failed to remove certificate: %w", err)
	}

	// Update ca-trust
	cmd := exec.Command("update-ca-trust", "extract")
	if s.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func (s *SystemStore) backupCaCertificates(backupPath string) error {
	// Backup /usr/local/share/ca-certificates/
	cmd := exec.Command("cp", "-r", "/usr/local/share/ca-certificates/", backupPath)
	return cmd.Run()
}

func (s *SystemStore) backupUpdateCaTrust(backupPath string) error {
	// Backup /etc/pki/ca-trust/source/anchors/
	cmd := exec.Command("cp", "-r", "/etc/pki/ca-trust/source/anchors/", backupPath)
	return cmd.Run()
}

func (s *SystemStore) restoreCaCertificates(backupPath string) error {
	// Restore /usr/local/share/ca-certificates/
	cmd := exec.Command("cp", "-r", backupPath, "/usr/local/share/ca-certificates/")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Update ca-certificates
	cmd = exec.Command("update-ca-certificates")
	return cmd.Run()
}

func (s *SystemStore) restoreUpdateCaTrust(backupPath string) error {
	// Restore /etc/pki/ca-trust/source/anchors/
	cmd := exec.Command("cp", "-r", backupPath, "/etc/pki/ca-trust/source/anchors/")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Update ca-trust
	cmd = exec.Command("update-ca-trust", "extract")
	return cmd.Run()
}

// Utility functions

func generateCertFilename(cert *x509.Certificate) string {
	// Generate a safe filename from certificate subject
	subject := cert.Subject.CommonName
	if subject == "" {
		subject = fmt.Sprintf("cert_%x", cert.SerialNumber)
	}

	// Replace unsafe characters
	filename := strings.ReplaceAll(subject, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "*", "_")

	return filename
}

func writeCertificateToFile(cert *x509.Certificate, path string) error {
	// Convert certificate to PEM format
	certPEM := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s-----END CERTIFICATE-----\n",
		base64.StdEncoding.EncodeToString(cert.Raw))

	// Write to file
	return os.WriteFile(path, []byte(certPEM), 0644)
}

// SupportedStores returns the list of supported stores for Linux
func SupportedStores() []string {
	var stores []string

	if _, err := exec.LookPath("update-ca-certificates"); err == nil {
		stores = append(stores, "ca-certificates")
	}

	if _, err := exec.LookPath("update-ca-trust"); err == nil {
		stores = append(stores, "update-ca-trust")
	}

	return stores
}
