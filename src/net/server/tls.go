package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	certFileName     = "cert.pem"
	keyFileName      = "key.pem"
	certValidityDays = 10 * 365
	serialBits       = 128
)

// LoadOrGenerateTLS returns a TLS server config using an existing certificate
// pair in dir, or generates a new self-signed ECDSA P-256 certificate if none
// exists. The certificate is valid for 10 years with SANs for localhost and
// 127.0.0.1.
func LoadOrGenerateTLS(dir string) (*tls.Config, error) {
	certPath := filepath.Join(dir, certFileName)
	keyPath := filepath.Join(dir, keyFileName)

	// Try loading existing pair first.
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err == nil {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		}, nil
	}

	// Generate new self-signed certificate.
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("create TLS directory: %w", err)
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialBits))
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{Organization: []string{"LootSheet"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(certValidityDays * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}, //nolint:mnd // IP literal
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	if err := os.WriteFile(certPath, certPEM, filePerm); err != nil {
		return nil, fmt.Errorf("write certificate: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, filePerm); err != nil {
		return nil, fmt.Errorf("write private key: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("load generated certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}
