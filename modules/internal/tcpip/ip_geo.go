// Package tcpip provides IP geolocation functionality
package tcpip

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// SimpleIPGeoDB is a simple IP geolocation database implementation
type SimpleIPGeoDB struct {
	// IPv4 range list
	ipv4Ranges []IPRange
	mu         sync.RWMutex
}

// IPRange represents an IP range to geolocation mapping
type IPRange struct {
	StartIP  uint32
	EndIP    uint32
	Country  string
	Region   string
	City     string
	ISP      string
	ASN      int
	Timezone string
}

// NewSimpleIPGeoDB creates a simple IP geolocation database
func NewSimpleIPGeoDB() *SimpleIPGeoDB {
	db := &SimpleIPGeoDB{
		ipv4Ranges: []IPRange{},
	}
	// Load default data for some common ranges
	db.loadDefaultRanges()
	return db
}

// Lookup queries IP geolocation
func (db *SimpleIPGeoDB) Lookup(ipStr string) (*GeoLocation, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Convert to IPv4
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("not an IPv4 address: %s", ipStr)
	}

	ipNum := ipToUint32(ip)

	db.mu.RLock()
	defer db.mu.RUnlock()

	// Linear search
	for _, r := range db.ipv4Ranges {
		if ipNum >= r.StartIP && ipNum <= r.EndIP {
			return &GeoLocation{
				Country:     r.Country,
				CountryCode: getCountryCode(r.Country),
				Region:      r.Region,
				City:        r.City,
				ISP:         r.ISP,
				ASN:         r.ASN,
				Timezone:    r.Timezone,
			}, nil
		}
	}

	return nil, fmt.Errorf("IP not found in database: %s", ipStr)
}

// GetRegionSignature gets the region signature
func (db *SimpleIPGeoDB) GetRegionSignature(country, isp string) string {
	return fmt.Sprintf("%s|%s", country, isp)
}

// LoadFromCSV loads data from a CSV file
func (db *SimpleIPGeoDB) LoadFromCSV(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// Skip header row
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 5 {
			continue
		}

		range_ := IPRange{
			Country:  record[2],
			Region:   record[3],
			City:     record[4],
			ISP:      getField(record, 5),
			Timezone: getField(record, 6),
		}

		// Parse IP range
		if strings.Contains(record[0], "-") {
			parts := strings.Split(record[0], "-")
			if len(parts) == 2 {
				range_.StartIP = ipStrToUint32(strings.TrimSpace(parts[0]))
				range_.EndIP = ipStrToUint32(strings.TrimSpace(parts[1]))
			}
		} else if strings.Contains(record[0], "/") {
			// CIDR format
			_, ipNet, err := net.ParseCIDR(record[0])
			if err == nil {
				range_.StartIP = ipToUint32(ipNet.IP)
				mask := binaryMask(ipNet.Mask)
				range_.EndIP = range_.StartIP | (^mask)
			}
		}

		if range_.StartIP != 0 && range_.EndIP != 0 {
			db.ipv4Ranges = append(db.ipv4Ranges, range_)
		}
	}

	return nil
}

// loadDefaultRanges loads default IP ranges
func (db *SimpleIPGeoDB) loadDefaultRanges() {
	// Copy defaults so runtime mutations do not alter the canonical list.
	db.ipv4Ranges = append(db.ipv4Ranges[:0], defaultIPRanges...)
}

// Helper functions

func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func ipStrToUint32(ipStr string) uint32 {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ipToUint32(ip)
}

func startIP(ipStr string) uint32 {
	return ipStrToUint32(ipStr)
}

func endIP(ipStr string) uint32 {
	return ipStrToUint32(ipStr)
}

func binaryMask(mask net.IPMask) uint32 {
	m := uint32(0)
	for i := 0; i < 4; i++ {
		m = m<<8 + uint32(mask[i])
	}
	return m
}

func getField(record []string, index int) string {
	if index < len(record) {
		return record[index]
	}
	return ""
}

func getCountryCode(country string) string {
	codes := map[string]string{
		"China":         "CN",
		"United States": "US",
		"Japan":         "JP",
		"Singapore":     "SG",
	}
	if code, ok := codes[country]; ok {
		return code
	}
	return ""
}
