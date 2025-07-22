package cert

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Fetcher handles fetching certificates from various sources
type Fetcher struct {
	httpClient *http.Client
	verbose    bool
}

// NewFetcher creates a new certificate fetcher
func NewFetcher(timeoutSeconds int, verbose bool) *Fetcher {
	return &Fetcher{
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		verbose: verbose,
	}
}

// FetchFromURL fetches certificates from a URL
func (f *Fetcher) FetchFromURL(url string, headers map[string]string, verifyTLS bool) ([]*x509.Certificate, error) {
	if f.verbose {
		fmt.Printf("Fetching certificates from URL: %s\n", url)
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Configure TLS verification
	if !verifyTLS {
		// This would require modifying the http client's transport
		// For now, we'll always verify TLS
	}

	// Make request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return f.ParseCertificates(data)
}

// FetchFromFile fetches certificates from a file
func (f *Fetcher) FetchFromFile(filePath string) ([]*x509.Certificate, error) {
	if f.verbose {
		fmt.Printf("Fetching certificates from file: %s\n", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return f.ParseCertificates(data)
}

// FetchFromDirectory fetches certificates from all files in a directory
func (f *Fetcher) FetchFromDirectory(dirPath string, filters []string) ([]*x509.Certificate, error) {
	if f.verbose {
		fmt.Printf("Fetching certificates from directory: %s\n", dirPath)
	}

	var allCerts []*x509.Certificate

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches filters
		if len(filters) > 0 && !matchesFilters(path, filters) {
			return nil
		}

		certs, err := f.FetchFromFile(path)
		if err != nil {
			if f.verbose {
				fmt.Printf("Warning: Failed to parse certificates from %s: %v\n", path, err)
			}
			return nil // Continue processing other files
		}

		allCerts = append(allCerts, certs...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allCerts, nil
}

// ParseCertificates parses certificates from PEM data
func (f *Fetcher) ParseCertificates(data []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	// Try to parse as PEM first
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				if f.verbose {
					fmt.Printf("Warning: Failed to parse certificate: %v\n", err)
				}
				continue
			}
			certs = append(certs, cert)
		}
	}

	// If no PEM certificates found, try to parse as DER
	if len(certs) == 0 {
		cert, err := x509.ParseCertificate(data)
		if err == nil {
			certs = append(certs, cert)
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no valid certificates found")
	}

	if f.verbose {
		fmt.Printf("Parsed %d certificates\n", len(certs))
	}

	return certs, nil
}

// ValidateCertificate validates a certificate
func (f *Fetcher) ValidateCertificate(cert *x509.Certificate) error {
	// Check if certificate is expired
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from %v)", cert.NotBefore)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired on %v)", cert.NotAfter)
	}

	// Check if it's a CA certificate
	if !cert.IsCA {
		return fmt.Errorf("certificate is not a CA certificate")
	}

	// Check basic constraints
	if !cert.BasicConstraintsValid {
		return fmt.Errorf("certificate has invalid basic constraints")
	}

	return nil
}

// GetCertificateFingerprint returns the SHA-256 fingerprint of a certificate
func GetCertificateFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(hash[:])
}

// GetCertificateInfo extracts information from a certificate
func GetCertificateInfo(cert *x509.Certificate) map[string]interface{} {
	return map[string]interface{}{
		"subject":       cert.Subject.String(),
		"issuer":        cert.Issuer.String(),
		"serial_number": cert.SerialNumber.String(),
		"not_before":    cert.NotBefore,
		"not_after":     cert.NotAfter,
		"fingerprint":   GetCertificateFingerprint(cert),
		"is_ca":         cert.IsCA,
		"key_usage":     cert.KeyUsage,
		"ext_key_usage": cert.ExtKeyUsage,
	}
}

// CompareCertificates compares two certificates for equality
func CompareCertificates(cert1, cert2 *x509.Certificate) bool {
	return GetCertificateFingerprint(cert1) == GetCertificateFingerprint(cert2)
}

// FilterCertificates filters certificates based on criteria
func FilterCertificates(certs []*x509.Certificate, filters []string) []*x509.Certificate {
	if len(filters) == 0 {
		return certs
	}

	var filtered []*x509.Certificate
	for _, cert := range certs {
		if matchesCertificateFilters(cert, filters) {
			filtered = append(filtered, cert)
		}
	}
	return filtered
}

// Helper functions

func matchesFilters(filePath string, filters []string) bool {
	fileName := filepath.Base(filePath)
	for _, filter := range filters {
		matched, _ := filepath.Match(filter, fileName)
		if matched {
			return true
		}
	}
	return false
}

func matchesCertificateFilters(cert *x509.Certificate, filters []string) bool {
	// For now, we'll implement basic subject matching
	subject := strings.ToLower(cert.Subject.String())
	for _, filter := range filters {
		if strings.Contains(subject, strings.ToLower(filter)) {
			return true
		}
	}
	return true // Default to include if no filters match
}

// ToPEM converts a certificate to PEM format
func ToPEM(cert *x509.Certificate) ([]byte, error) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(block), nil
}
