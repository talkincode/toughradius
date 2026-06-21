package app

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// CertStore resolves locally managed certificates (domain.SysCert) by their
// operator-assigned name. It backs the EAP-TLS/PEAP/TTLS providers so a
// certificate selected on the system config page (radius.EapTlsServerCert /
// radius.EapTlsClientCa) is loaded from the database in preference to on-disk
// file paths.
//
// CertStore satisfies the eap/handlers.CertResolver interface structurally,
// which keeps the RADIUS handlers package free of an import dependency on
// internal/app and the domain models while still letting it load managed
// certificates through this type.
type CertStore struct {
	db *gorm.DB
}

// NewCertStore returns a CertStore backed by the given database handle.
func NewCertStore(db *gorm.DB) *CertStore {
	return &CertStore{db: db}
}

// ServerKeyPair returns the PEM certificate chain and PEM private key for the
// managed server certificate with the given name. It returns an error when the
// certificate is unknown or has no stored private key, so the caller can surface
// a misconfiguration rather than silently fall back to an unauthenticated state.
func (s *CertStore) ServerKeyPair(name string) (certPEM, keyPEM []byte, err error) {
	if s == nil || s.db == nil {
		return nil, nil, fmt.Errorf("certificate store not initialized")
	}
	var cert domain.SysCert
	if err := s.db.Where("name = ?", name).First(&cert).Error; err != nil {
		return nil, nil, fmt.Errorf("load managed certificate %q: %w", name, err)
	}
	if cert.PrivateKey == "" {
		return nil, nil, fmt.Errorf("managed certificate %q has no stored private key", name)
	}
	return []byte(cert.Cert), []byte(cert.PrivateKey), nil
}

// CABundle returns the PEM CA bundle for the managed certificate with the given
// name. It returns an error when the certificate is unknown.
func (s *CertStore) CABundle(name string) (caPEM []byte, err error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("certificate store not initialized")
	}
	var cert domain.SysCert
	if err := s.db.Where("name = ?", name).First(&cert).Error; err != nil {
		return nil, fmt.Errorf("load managed CA certificate %q: %w", name, err)
	}
	return []byte(cert.Cert), nil
}
