package ja4x

import (
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"strings"
)

// translated comment
type JA4XResult struct {
	// translated comment
	Hash string

	// translated comment
	RawString string

	// translated comment
	JA4Xa string

	// translated comment
	JA4Xb string

	// translated comment
	JA4Xc string

	// translated comment
	Issuer string

	// translated comment
	Subject string

	// translated comment
	AnomalyFlags []string

	// translated comment
	RiskScore float64
}

// translated comment
type CertificateData struct {
	// translated comment
	IssuerRDNs []RDNField

	// translated comment
	SubjectRDNs []RDNField

	// translated comment
	ExtensionOIDs []string

	// translated comment
	Version int

	// translated comment
	SignatureAlgorithm string

	// translated comment
	ValidityDays int

	// translated comment
	IsSelfSigned bool

	// translated comment
	IsCA bool
}

// translated comment
type RDNField struct {
	// translated comment
	OID string

	// translated comment
	Value string
}

// translated comment
func sha256Hash12(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)[:12]
}

// translated comment
func ComputeJA4X(cert *x509.Certificate) *JA4XResult {
	data := extractCertificateData(cert)
	return ComputeJA4XFromData(data)
}

// translated comment
func ComputeJA4XFromData(data CertificateData) *JA4XResult {
	result := &JA4XResult{
		AnomalyFlags: []string{},
	}

	// translated comment
	issuerOIDs := make([]string, len(data.IssuerRDNs))
	for i, rdn := range data.IssuerRDNs {
		issuerOIDs[i] = rdn.OID
	}
	issuerStr := strings.Join(issuerOIDs, ",")
	result.JA4Xa = sha256Hash12(issuerStr)

	// translated comment
	issuerParts := make([]string, len(data.IssuerRDNs))
	for i, rdn := range data.IssuerRDNs {
		issuerParts[i] = fmt.Sprintf("%s=%s", oidName(rdn.OID), rdn.Value)
	}
	result.Issuer = strings.Join(issuerParts, ", ")

	// translated comment
	subjectOIDs := make([]string, len(data.SubjectRDNs))
	for i, rdn := range data.SubjectRDNs {
		subjectOIDs[i] = rdn.OID
	}
	subjectStr := strings.Join(subjectOIDs, ",")
	result.JA4Xb = sha256Hash12(subjectStr)

	// translated comment
	subjectParts := make([]string, len(data.SubjectRDNs))
	for i, rdn := range data.SubjectRDNs {
		subjectParts[i] = fmt.Sprintf("%s=%s", oidName(rdn.OID), rdn.Value)
	}
	result.Subject = strings.Join(subjectParts, ", ")

	// translated comment
	extStr := strings.Join(data.ExtensionOIDs, ",")
	result.JA4Xc = sha256Hash12(extStr)

	// translated comment
	result.RawString = fmt.Sprintf("%s_%s_%s", issuerStr, subjectStr, extStr)

	// translated comment
	result.Hash = fmt.Sprintf("%s_%s_%s", result.JA4Xa, result.JA4Xb, result.JA4Xc)

	// translated comment
	detectCertAnomalies(data, result)

	return result
}

// translated comment
func ComputeJA4XChain(certs []*x509.Certificate) []*JA4XResult {
	results := make([]*JA4XResult, len(certs))
	for i, cert := range certs {
		results[i] = ComputeJA4X(cert)
	}
	return results
}

// translated comment
func MatchJA4X(hash1, hash2 string) bool {
	return hash1 != "" && hash2 != "" && hash1 == hash2
}

// translated comment
func extractCertificateData(cert *x509.Certificate) CertificateData {
	data := CertificateData{
		Version:      cert.Version,
		IsSelfSigned: cert.Issuer.String() == cert.Subject.String(),
		IsCA:         cert.IsCA,
	}

	// translated comment
	data.SignatureAlgorithm = cert.SignatureAlgorithm.String()

	// translated comment
	if !cert.NotBefore.IsZero() && !cert.NotAfter.IsZero() {
		data.ValidityDays = int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)
	}

	// translated comment
	data.IssuerRDNs = extractRDNs(cert.Issuer.String())

	// translated comment
	data.SubjectRDNs = extractRDNs(cert.Subject.String())

	// translated comment
	for _, ext := range cert.Extensions {
		data.ExtensionOIDs = append(data.ExtensionOIDs, ext.Id.String())
	}

	return data
}

// translated comment
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

// translated comment
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
		// translated comment
		if strings.Contains(key, ".") {
			return key
		}
		return key
	}
}

// translated comment
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

// translated comment
func detectCertAnomalies(data CertificateData, result *JA4XResult) {
	baseScore := 0.0

	// translated comment
	if data.IsSelfSigned {
		result.AnomalyFlags = append(result.AnomalyFlags, "SELF_SIGNED")
		baseScore += 0.2
	}

	// translated comment
	if data.ValidityDays > 398 && !data.IsCA {
		result.AnomalyFlags = append(result.AnomalyFlags, "LONG_VALIDITY")
		baseScore += 0.15
	}

	// translated comment
	if data.Version != 3 && data.Version != 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NON_V3_CERT")
		baseScore += 0.1
	}

	// translated comment
	if len(data.ExtensionOIDs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_EXTENSIONS")
		baseScore += 0.15
	}

	// translated comment
	if len(data.SubjectRDNs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_SUBJECT")
		baseScore += 0.1
	}

	// translated comment
	if len(data.IssuerRDNs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_ISSUER")
		baseScore += 0.15
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}
