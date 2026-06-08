package tlsengine

import (
	"crypto/x509"
	"strings"
)

// PeerIdentity is the validated identity of an EAP-TLS peer, derived from its
// (already CA-verified) client certificate per RFC 5216 §5.2.
type PeerIdentity struct {
	// Name is the primary peer identity (Peer-Id). It is taken from the
	// subjectAltName when present (rfc822Name preferred, then dnsName), and
	// otherwise from the subject CommonName.
	Name string
	// Source records where Name was derived from ("san-rfc822", "san-dns" or
	// "subject-cn"), for diagnostics and identity-binding decisions.
	Source string
	// SANEmails / SANDNSNames expose all subjectAltName entries so callers can
	// implement custom identity-matching policies.
	SANEmails   []string
	SANDNSNames []string
	// CommonName is the subject CommonName, if any.
	CommonName string
}

// Identity-derivation source labels.
const (
	SourceSANEmail = "san-rfc822"
	SourceSANDNS   = "san-dns"
	SourceSubject  = "subject-cn"
)

// identityFromCertificate determines the Peer-Id from a verified client
// certificate following RFC 5216 §5.2: where a subjectAltName is present it
// takes precedence (an rfc822Name is preferred for a user identity, then a
// dnsName); otherwise the subject CommonName is used.
func identityFromCertificate(cert *x509.Certificate) *PeerIdentity {
	id := &PeerIdentity{
		SANEmails:   cert.EmailAddresses,
		SANDNSNames: cert.DNSNames,
		CommonName:  cert.Subject.CommonName,
	}

	switch {
	case len(cert.EmailAddresses) > 0 && strings.TrimSpace(cert.EmailAddresses[0]) != "":
		id.Name = strings.TrimSpace(cert.EmailAddresses[0])
		id.Source = SourceSANEmail
	case len(cert.DNSNames) > 0 && strings.TrimSpace(cert.DNSNames[0]) != "":
		id.Name = strings.TrimSpace(cert.DNSNames[0])
		id.Source = SourceSANDNS
	default:
		id.Name = strings.TrimSpace(cert.Subject.CommonName)
		id.Source = SourceSubject
	}
	return id
}

// Matches reports whether the supplied username corresponds to this peer
// identity. The comparison is case-insensitive and considers the primary
// identity Name, the subject CommonName, and every subjectAltName entry, so an
// EAP-Response/Identity that carries any of the certificate's names is accepted.
// An empty username never matches.
func (p *PeerIdentity) Matches(username string) bool {
	want := strings.TrimSpace(username)
	if want == "" {
		return false
	}
	candidates := make([]string, 0, len(p.SANEmails)+len(p.SANDNSNames)+2)
	candidates = append(candidates, p.Name, p.CommonName)
	candidates = append(candidates, p.SANEmails...)
	candidates = append(candidates, p.SANDNSNames...)
	for _, c := range candidates {
		if strings.EqualFold(strings.TrimSpace(c), want) {
			return true
		}
	}
	return false
}
