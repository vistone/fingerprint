package types

import (
	"testing"
)

func TestBrowserTypeConstants(t *testing.T) {
	// Test that browser type constants are defined
	browsers := []BrowserType{
		BrowserChrome,
		BrowserFirefox,
		BrowserSafari,
		BrowserOpera,
		BrowserEdge,
	}

	for _, browser := range browsers {
		if browser == "" {
			t.Error("Browser type should not be empty")
		}
	}
}

func TestOperatingSystemConstants(t *testing.T) {
	// Test that OS constants are defined
	oses := []OperatingSystem{
		OSWindows10,
		OSWindows11,
		OSMacOS13,
		OSMacOS14,
		OSMacOS15,
		OSLinux,
		OSLinuxUbuntu,
		OSLinuxDebian,
	}

	for _, os := range oses {
		if os == "" {
			t.Error("Operating system should not be empty")
		}
	}
}

func TestOperatingSystemsSlice(t *testing.T) {
	if len(OperatingSystems) == 0 {
		t.Error("OperatingSystems slice should not be empty")
	}
}

func TestFingerprintResult(t *testing.T) {
	// Test that FingerprintResult can be created
	result := FingerprintResult{
		UserAgent:     "Mozilla/5.0",
		HelloClientID: "chrome_120",
	}

	if result.UserAgent != "Mozilla/5.0" {
		t.Error("UserAgent mismatch")
	}
}
