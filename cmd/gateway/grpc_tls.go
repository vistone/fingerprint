package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

var (
	errGRPCTLSNotConfigured   = errors.New("gRPC TLS is not configured")
	errGRPCTLSCertKeyRequired = errors.New("both FP_GRPC_TLS_CERT_FILE and FP_GRPC_TLS_KEY_FILE are required")
	errGRPCTLSParseCAPEM      = errors.New("parse CA PEM failed")
	errNoTLSCertificate       = errors.New("no TLS certificate configured")
	errNoTLSCertificateChain  = errors.New("no TLS certificate chain configured")
)

func buildGRPCTLSConfigFromEnv() (*tls.Config, error) {
	certPath, certOK := readEnvString("FP_GRPC_TLS_CERT_FILE")
	keyPath, keyOK := readEnvString("FP_GRPC_TLS_KEY_FILE")
	caPath, caOK := readEnvString("FP_GRPC_TLS_CA_FILE")

	if !certOK && !keyOK && !caOK {
		return nil, errGRPCTLSNotConfigured
	}

	if !certOK || !keyOK {
		return nil, errGRPCTLSCertKeyRequired
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("load server cert/key failed: %w", err)
	}

	tlsConfig := new(tls.Config)
	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.Certificates = []tls.Certificate{cert}

	if caOK {
		caData, readErr := os.ReadFile(caPath)
		if readErr != nil {
			return nil, fmt.Errorf("read CA file failed: %w", readErr)
		}

		caPool := x509.NewCertPool()

		if !caPool.AppendCertsFromPEM(caData) {
			return nil, errGRPCTLSParseCAPEM
		}

		tlsConfig.ClientCAs = caPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// Validate config early to fail fast.
	err = validateTLSConfig(tlsConfig)
	if err != nil {
		return nil, err
	}

	return tlsConfig, nil
}

func validateTLSConfig(cfg *tls.Config) error {
	if cfg == nil {
		return nil
	}

	if len(cfg.Certificates) == 0 {
		return errNoTLSCertificate
	}

	// Ensure at least one cert has parsed leaf; parse if needed.
	for certIndex := range cfg.Certificates {
		if len(cfg.Certificates[certIndex].Certificate) == 0 {
			continue
		}

		if cfg.Certificates[certIndex].Leaf != nil {
			return nil
		}

		leaf, err := x509.ParseCertificate(cfg.Certificates[certIndex].Certificate[0])
		if err != nil {
			return fmt.Errorf("parse TLS leaf certificate failed: %w", err)
		}

		cfg.Certificates[certIndex].Leaf = leaf

		return nil
	}

	return errNoTLSCertificateChain
}
