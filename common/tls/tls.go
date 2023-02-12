package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

type Config struct {
	Country            []string
	OrganizationalUnit []string
	Organization       []string
	Locality           []string
	Province           []string
	StreetAddress      []string
	PostalCode         []string
	CommonName         string
	SerialNumber       string
	DNSNames           []string
	Years              int
}

func GeneratePrivateKey() (key *ecdsa.PrivateKey) {
	key, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return
}

func GenerateCaCrt(config Config, crtpath, keypath string) error {
	rootkey := GeneratePrivateKey()
	var rootCsr = &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            config.Country,
			Province:           config.Province,
			Locality:           config.Locality,
			Organization:       config.Organization,
			OrganizationalUnit: config.OrganizationalUnit,
			CommonName:         config.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(config.Years, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	rootDer, err := x509.CreateCertificate(rand.Reader, rootCsr, rootCsr, rootkey.Public(), rootkey)
	if err != nil {
		return err
	}
	rootCert, err := x509.ParseCertificate(rootDer)

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rootCert.Raw,
	}

	pemData := pem.EncodeToMemory(certBlock)
	if err = os.WriteFile(crtpath, pemData, 0644); err != nil {
		return err
	}

	keyDer, err := x509.MarshalECPrivateKey(rootkey)
	if err != nil {
		return err
	}

	keyBlock := &pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: keyDer,
	}
	keyData := pem.EncodeToMemory(keyBlock)
	if err = os.WriteFile(keypath, keyData, 0644); err != nil {
		return err
	}
	return nil

}

func GenerateCrt(config Config, cacrtfile, cakeyfile, crtfile, keyfile string) error {
	cadata, err := tls.LoadX509KeyPair(cacrtfile, cakeyfile)
	if err != nil {
		return err
	}
	cacrt, err := x509.ParseCertificate(cadata.Certificate[0])
	if err != nil {
		return err
	}
	cakbytes := cadata.PrivateKey.(*ecdsa.PrivateKey).D.Bytes()
	capkey, err := x509.ParseECPrivateKey(cakbytes)
	if err != nil {
		return err
	}

	var csr = &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            config.Country,
			Province:           config.Province,
			Locality:           config.Locality,
			Organization:       config.Organization,
			OrganizationalUnit: config.OrganizationalUnit,
			CommonName:         config.CommonName,
		},
		DNSNames:              config.DNSNames,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(config.Years, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	key := GeneratePrivateKey()
	der, err := x509.CreateCertificate(rand.Reader, csr, cacrt, key.Public(), capkey)
	if err != nil {
		return err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return err
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	pemData := pem.EncodeToMemory(certBlock)
	if err = os.WriteFile(crtfile, pemData, 0644); err != nil {
		return err
	}

	keyDer, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	keyBlock := &pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: keyDer,
	}

	keyData := pem.EncodeToMemory(keyBlock)

	if err = os.WriteFile(keyfile, keyData, 0644); err != nil {
		return err
	}

	return nil

}
