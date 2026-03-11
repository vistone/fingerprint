package ja4x

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// generateTestCert generates a test certificate
func generateTestCert(t *testing.T, template, parent *x509.Certificate, pub, priv interface{}) *x509.Certificate {
	t.Helper()
	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}
	return cert
}

func TestComputeJA4X(t *testing.T) {
	// Generate test key pair
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	t.Run("self_signed_cert", func(t *testing.T) {
		template := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName:   "test.example.com",
				Organization: []string{"Test Org"},
				Country:      []string{"US"},
			},
			Issuer: pkix.Name{
				CommonName:   "test.example.com",
				Organization: []string{"Test Org"},
				Country:      []string{"US"},
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			BasicConstraintsValid: true,
		}

		cert := generateTestCert(t, template, template, &key.PublicKey, key)
		result := ComputeJA4X(cert)

		if result == nil {
			t.Fatal("ComputeJA4X returned nil")
		}
		if result.Hash == "" {
			t.Error("Hash should not be empty")
		}
		if result.JA4Xa == "" {
			t.Error("JA4Xa should not be empty")
		}
		if result.JA4Xb == "" {
			t.Error("JA4Xb should not be empty")
		}
		if result.JA4Xc == "" {
			t.Error("JA4Xc should not be empty")
		}

		// Self-signed certificate JA4Xa and JA4Xb should be the same
		if result.JA4Xa != result.JA4Xb {
			t.Errorf("Self-signed cert should have JA4Xa == JA4Xb, got %s != %s", result.JA4Xa, result.JA4Xb)
		}

		// Check anomaly flags
		foundSelfSigned := false
		for _, flag := range result.AnomalyFlags {
			if flag == "SELF_SIGNED" {
				foundSelfSigned = true
			}
		}
		if !foundSelfSigned {
			t.Error("Expected SELF_SIGNED anomaly flag")
		}
	})

	t.Run("ca_cert", func(t *testing.T) {
		template := &x509.Certificate{
			SerialNumber: big.NewInt(2),
			Subject: pkix.Name{
				CommonName:   "Test CA",
				Organization: []string{"Test CA Org"},
				Country:      []string{"US"},
			},
			Issuer: pkix.Name{
				CommonName:   "Test CA",
				Organization: []string{"Test CA Org"},
				Country:      []string{"US"},
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(3650 * 24 * time.Hour),
			IsCA:                  true,
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			BasicConstraintsValid: true,
		}

		cert := generateTestCert(t, template, template, &key.PublicKey, key)
		result := ComputeJA4X(cert)

		// CA cert with long validity should NOT flag LONG_VALIDITY
		for _, flag := range result.AnomalyFlags {
			if flag == "LONG_VALIDITY" {
				t.Error("CA cert should not flag LONG_VALIDITY")
			}
		}
	})
}

func TestComputeJA4XFromData(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		data := CertificateData{
			IssuerRDNs: []RDNField{
				{OID: "2.5.4.6", Value: "US"},
				{OID: "2.5.4.10", Value: "Let's Encrypt"},
				{OID: "2.5.4.3", Value: "R3"},
			},
			SubjectRDNs: []RDNField{
				{OID: "2.5.4.3", Value: "example.com"},
			},
			ExtensionOIDs: []string{
				"2.5.29.15",  // Key Usage
				"2.5.29.37",  // Extended Key Usage
				"2.5.29.14",  // Subject Key Identifier
				"2.5.29.35",  // Authority Key Identifier
				"2.5.29.17",  // Subject Alternative Name
				"1.3.6.1.5.5.7.1.1", // Authority Info Access
			},
			Version:            3,
			SignatureAlgorithm: "SHA256-RSA",
			ValidityDays:       90,
			IsSelfSigned:       false,
			IsCA:               false,
		}

		result := ComputeJA4XFromData(data)
		if result == nil {
			t.Fatal("ComputeJA4XFromData returned nil")
		}
		if len(result.JA4Xa) != 12 {
			t.Errorf("JA4Xa length = %d, want 12", len(result.JA4Xa))
		}
		if len(result.JA4Xb) != 12 {
			t.Errorf("JA4Xb length = %d, want 12", len(result.JA4Xb))
		}
		if len(result.JA4Xc) != 12 {
			t.Errorf("JA4Xc length = %d, want 12", len(result.JA4Xc))
		}

		// Issuer and Subject should be different
		if result.JA4Xa == result.JA4Xb {
			t.Error("Different issuer and subject should produce different JA4X_a and JA4X_b")
		}

		// Risk score should be 0 for a normal cert
		if result.RiskScore != 0 {
			t.Errorf("RiskScore = %f, want 0 for normal cert", result.RiskScore)
		}
	})

	t.Run("anomalous_cert", func(t *testing.T) {
		data := CertificateData{
			IssuerRDNs:    []RDNField{},
			SubjectRDNs:   []RDNField{},
			ExtensionOIDs: []string{},
			Version:       1,
			ValidityDays:  730,
			IsSelfSigned:  true,
			IsCA:          false,
		}

		result := ComputeJA4XFromData(data)
		if result.RiskScore == 0 {
			t.Error("Expected non-zero risk score for anomalous cert")
		}

		expectedFlags := map[string]bool{
			"SELF_SIGNED":   true,
			"LONG_VALIDITY": true,
			"NO_EXTENSIONS": true,
			"EMPTY_SUBJECT": true,
			"EMPTY_ISSUER":  true,
		}

		for _, flag := range result.AnomalyFlags {
			delete(expectedFlags, flag)
		}
		for flag := range expectedFlags {
			t.Errorf("Expected anomaly flag %s not found", flag)
		}
	})
}

func TestMatchJA4X(t *testing.T) {
	data := CertificateData{
		IssuerRDNs: []RDNField{
			{OID: "2.5.4.3", Value: "Test CA"},
		},
		SubjectRDNs: []RDNField{
			{OID: "2.5.4.3", Value: "test.com"},
		},
		ExtensionOIDs: []string{"2.5.29.15"},
	}

	r1 := ComputeJA4XFromData(data)
	r2 := ComputeJA4XFromData(data)

	if !MatchJA4X(r1.Hash, r2.Hash) {
		t.Error("Same data should produce matching hashes")
	}

	data2 := CertificateData{
		IssuerRDNs: []RDNField{
			{OID: "2.5.4.3", Value: "Different CA"},
		},
		SubjectRDNs: []RDNField{
			{OID: "2.5.4.3", Value: "other.com"},
		},
		ExtensionOIDs: []string{"2.5.29.15", "2.5.29.37"},
	}

	r3 := ComputeJA4XFromData(data2)
	if MatchJA4X(r1.Hash, r3.Hash) {
		t.Error("Different data should produce different hashes")
	}

	if MatchJA4X("", r1.Hash) {
		t.Error("Empty hash should not match")
	}
}

func TestComputeJA4XChain(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	// CA cert
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test CA Org"},
		},
		Issuer: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test CA Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(3650 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caCert := generateTestCert(t, caTemplate, caTemplate, &key.PublicKey, key)

	// Leaf cert
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: "leaf.example.com",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}
	leafCert := generateTestCert(t, leafTemplate, caTemplate, &key.PublicKey, key)

	results := ComputeJA4XChain([]*x509.Certificate{leafCert, caCert})
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	// Leaf and CA should have different fingerprints
	if results[0].Hash == results[1].Hash {
		t.Error("Leaf and CA certs should have different JA4X fingerprints")
	}
}

func TestRDNKeyToOID(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"CN", "2.5.4.3"},
		{"O", "2.5.4.10"},
		{"OU", "2.5.4.11"},
		{"C", "2.5.4.6"},
		{"ST", "2.5.4.8"},
		{"L", "2.5.4.7"},
		{"2.5.4.3", "2.5.4.3"}, // Already OID format
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		got := rdnKeyToOID(tt.key)
		if got != tt.want {
			t.Errorf("rdnKeyToOID(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestOIDName(t *testing.T) {
	tests := []struct {
		oid  string
		want string
	}{
		{"2.5.4.3", "CN"},
		{"2.5.4.10", "O"},
		{"2.5.4.6", "C"},
		{"1.2.3.4", "1.2.3.4"}, // Unknown OID
	}

	for _, tt := range tests {
		got := oidName(tt.oid)
		if got != tt.want {
			t.Errorf("oidName(%q) = %q, want %q", tt.oid, got, tt.want)
		}
	}
}

func TestDeterministicHash(t *testing.T) {
	data := CertificateData{
		IssuerRDNs: []RDNField{
			{OID: "2.5.4.6", Value: "US"},
			{OID: "2.5.4.3", Value: "Test"},
		},
		SubjectRDNs: []RDNField{
			{OID: "2.5.4.3", Value: "example.com"},
		},
		ExtensionOIDs: []string{"2.5.29.15", "2.5.29.37"},
	}

	r1 := ComputeJA4XFromData(data)
	r2 := ComputeJA4XFromData(data)

	if r1.Hash != r2.Hash {
		t.Error("Same input should produce same hash")
	}
	if r1.RawString != r2.RawString {
		t.Error("Same input should produce same raw string")
	}
}

// BenchmarkComputeJA4X benchmark
func BenchmarkComputeJA4X(b *testing.B) {
	data := CertificateData{
		IssuerRDNs: []RDNField{
			{OID: "2.5.4.6", Value: "US"},
			{OID: "2.5.4.10", Value: "Let's Encrypt"},
			{OID: "2.5.4.3", Value: "R3"},
		},
		SubjectRDNs: []RDNField{
			{OID: "2.5.4.3", Value: "example.com"},
		},
		ExtensionOIDs: []string{
			"2.5.29.15", "2.5.29.37", "2.5.29.14",
			"2.5.29.35", "2.5.29.17", "1.3.6.1.5.5.7.1.1",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ComputeJA4XFromData(data)
	}
}
