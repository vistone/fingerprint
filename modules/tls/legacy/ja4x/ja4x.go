package ja4x

import (
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"strings"
)

// JA4XResult JA4X fingerprint result (X.509 certificate fingerprint)
type JA4XResult struct {
	// JA4X full hash
	Hash string

	// JA4X raw string
	RawString string

	// JA4X_a: first 12 characters of SHA256 of issuer RDN (Distinguished Name) sequence
	JA4Xa string

	// JA4X_b: first 12 characters of SHA256 of subject RDN sequence
	JA4Xb string

	// JA4X_c: first 12 characters of SHA256 of extension OID sequence
	JA4Xc string

	// Issuer info (human-readable form)
	Issuer string

	// Subject info (human-readable form)
	Subject string

	// Anomaly flags
	AnomalyFlags []string

	// Risk score (0.0-1.0)
	RiskScore float64
}

// CertificateData certificate data (for scenarios not depending on x509.Certificate)
type CertificateData struct {
	// Issuer RDN fields (in order)
	IssuerRDNs []RDNField

	// Subject RDN fields (in order)
	SubjectRDNs []RDNField

	// Extension OID list (in order)
	ExtensionOIDs []string

	// Certificate version
	Version int

	// Signature algorithm OID
	SignatureAlgorithm string

	// Validity period (days)
	ValidityDays int

	// Whether self-signed
	IsSelfSigned bool

	// Whether it is a CA certificate
	IsCA bool
}

// RDNField certificate RDN field
type RDNField struct {
	// OID (e.g. "2.5.4.3" for CN)
	OID string

	// Value
	Value string
}

// sha256Hash12 computes SHA256 hash and returns the first 12 characters
func sha256Hash12(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)[:12]
}

// ComputeJA4X computes JA4X fingerprint from x509.Certificate
func ComputeJA4X(cert *x509.Certificate) *JA4XResult {
	data := extractCertificateData(cert)
	return ComputeJA4XFromData(data)
}

// ComputeJA4XFromData computes JA4X fingerprint from CertificateData
func ComputeJA4XFromData(data CertificateData) *JA4XResult {
	result := &JA4XResult{
		AnomalyFlags: []string{},
	}

	// JA4X_a: issuer RDN OID sequence
	issuerOIDs := make([]string, len(data.IssuerRDNs))
	for i, rdn := range data.IssuerRDNs {
		issuerOIDs[i] = rdn.OID
	}
	issuerStr := strings.Join(issuerOIDs, ",")
	result.JA4Xa = sha256Hash12(issuerStr)

	// Issuer human-readable form
	issuerParts := make([]string, len(data.IssuerRDNs))
	for i, rdn := range data.IssuerRDNs {
		issuerParts[i] = fmt.Sprintf("%s=%s", oidName(rdn.OID), rdn.Value)
	}
	result.Issuer = strings.Join(issuerParts, ", ")

	// JA4X_b: subject RDN OID sequence
	subjectOIDs := make([]string, len(data.SubjectRDNs))
	for i, rdn := range data.SubjectRDNs {
		subjectOIDs[i] = rdn.OID
	}
	subjectStr := strings.Join(subjectOIDs, ",")
	result.JA4Xb = sha256Hash12(subjectStr)

	// Subject human-readable form
	subjectParts := make([]string, len(data.SubjectRDNs))
	for i, rdn := range data.SubjectRDNs {
		subjectParts[i] = fmt.Sprintf("%s=%s", oidName(rdn.OID), rdn.Value)
	}
	result.Subject = strings.Join(subjectParts, ", ")

	// JA4X_c: extension OID sequence
	extStr := strings.Join(data.ExtensionOIDs, ",")
	result.JA4Xc = sha256Hash12(extStr)

	// Raw string: issuerOIDs_subjectOIDs_extensionOIDs
	result.RawString = fmt.Sprintf("%s_%s_%s", issuerStr, subjectStr, extStr)

	// Full hash: JA4X_a_JA4X_b_JA4X_c
	result.Hash = fmt.Sprintf("%s_%s_%s", result.JA4Xa, result.JA4Xb, result.JA4Xc)

	// Anomaly detection
	detectCertAnomalies(data, result)

	return result
}

// ComputeJA4XChain computes JA4X fingerprint list for certificate chain
func ComputeJA4XChain(certs []*x509.Certificate) []*JA4XResult {
	results := make([]*JA4XResult, len(certs))
	for i, cert := range certs {
		results[i] = ComputeJA4X(cert)
	}
	return results
}

// MatchJA4X compares whether two JA4X hashes match
func MatchJA4X(hash1, hash2 string) bool {
	return hash1 != "" && hash2 != "" && hash1 == hash2
}

// extractCertificateData extracts data from x509.Certificate
func extractCertificateData(cert *x509.Certificate) CertificateData {
	data := CertificateData{
		Version:      cert.Version,
		IsSelfSigned: cert.Issuer.String() == cert.Subject.String(),
		IsCA:         cert.IsCA,
	}

	// Signature algorithm
	data.SignatureAlgorithm = cert.SignatureAlgorithm.String()

	// Validity period
	if !cert.NotBefore.IsZero() && !cert.NotAfter.IsZero() {
		data.ValidityDays = int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)
	}

	// Issuer RDN
	data.IssuerRDNs = extractRDNs(cert.Issuer.String())

	// Subject RDN
	data.SubjectRDNs = extractRDNs(cert.Subject.String())

	// Extension OID
	for _, ext := range cert.Extensions {
		data.ExtensionOIDs = append(data.ExtensionOIDs, ext.Id.String())
	}

	return data
}

// extractRDNs extracts RDN fields from DN string
func extractRDNs(dn string) []RDNField {
	var rdns []RDNField
	if dn == "" {
		return rdns
	}

	parts := strings.Split(dn, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		eqIdx := strings.Index(part, "=")
		if eqIdx > 0 {
			key := strings.TrimSpace(part[:eqIdx])
			value := strings.TrimSpace(part[eqIdx+1:])
			rdns = append(rdns, RDNField{
				OID:   rdnKeyToOID(key),
				Value: value,
			})
		}
	}
	return rdns
}

// rdnKeyToOID converts RDN key name to OID
func rdnKeyToOID(key string) string {
	switch strings.ToUpper(key) {
	case "CN":
		return "2.5.4.3"
	case "O":
		return "2.5.4.10"
	case "OU":
		return "2.5.4.11"
	case "C":
		return "2.5.4.6"
	case "ST", "S":
		return "2.5.4.8"
	case "L":
		return "2.5.4.7"
	case "SERIALNUMBER":
		return "2.5.4.5"
	case "STREET":
		return "2.5.4.9"
	case "POSTALCODE":
		return "2.5.4.17"
	default:
		// If already in OID format, return directly
		if strings.Contains(key, ".") {
			return key
		}
		return key
	}
}

// oidName converts OID to human-readable name
func oidName(oid string) string {
	switch oid {
	case "2.5.4.3":
		return "CN"
	case "2.5.4.10":
		return "O"
	case "2.5.4.11":
		return "OU"
	case "2.5.4.6":
		return "C"
	case "2.5.4.8":
		return "ST"
	case "2.5.4.7":
		return "L"
	case "2.5.4.5":
		return "SERIALNUMBER"
	case "2.5.4.9":
		return "STREET"
	case "2.5.4.17":
		return "POSTALCODE"
	default:
		return oid
	}
}

// detectCertAnomalies detects certificate anomalies
func detectCertAnomalies(data CertificateData, result *JA4XResult) {
	baseScore := 0.0

	// Anomaly 1: self-signed certificate
	if data.IsSelfSigned {
		result.AnomalyFlags = append(result.AnomalyFlags, "SELF_SIGNED")
		baseScore += 0.2
	}

	// Anomaly 2: excessive validity period (over 398 days, violates CA/Browser Forum requirements)
	if data.ValidityDays > 398 && !data.IsCA {
		result.AnomalyFlags = append(result.AnomalyFlags, "LONG_VALIDITY")
		baseScore += 0.15
	}

	// Anomaly 3: certificate version is not v3
	if data.Version != 3 && data.Version != 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NON_V3_CERT")
		baseScore += 0.1
	}

	// Anomaly 4: no extensions
	if len(data.ExtensionOIDs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_EXTENSIONS")
		baseScore += 0.15
	}

	// Anomaly 5: empty subject
	if len(data.SubjectRDNs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_SUBJECT")
		baseScore += 0.1
	}

	// Anomaly 6: empty issuer
	if len(data.IssuerRDNs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_ISSUER")
		baseScore += 0.15
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}
