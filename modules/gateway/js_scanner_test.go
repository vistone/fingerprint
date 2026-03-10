package gateway

import (
	"context"
	"errors"
	"testing"
)

func findDetectedByName(result *JSDetectionResult, name string) *DetectedAPI {
	for i := range result.Detected {
		if result.Detected[i].Name == name {
			return &result.Detected[i]
		}
	}
	return nil
}

func TestScanJavaScriptWithV8_DetectsCapturedScriptSrc(t *testing.T) {
	html := `
	<html>
	  <body>
	    <script data-captured-src="https://www.googletagmanager.com/gtag/js?id=DC-8128335&amp;l=marketingClientDataLayer"></script>
	  </body>
	</html>
	`

	result, err := ScanJavaScriptWithV8(context.Background(), html)
	if err != nil {
		t.Fatalf("ScanJavaScriptWithV8 returned error: %v", err)
	}

	hit := findDetectedByName(result, "Google Analytics Tracking")
	if hit == nil {
		t.Fatalf("expected Google Analytics Tracking to be detected, got: %+v", result.Detected)
	}
	if hit.Count <= 0 {
		t.Fatalf("expected positive match count, got %d", hit.Count)
	}
}

func TestScanJavaScriptWithV8_DeduplicatesRepeatedMatches(t *testing.T) {
	html := `
	<html>
	  <body>
	    <script>
	      var a = "_ga";
	      var b = "_ga";
	      var c = "_ga";
	    </script>
	  </body>
	</html>
	`

	result, err := ScanJavaScriptWithV8(context.Background(), html)
	if err != nil {
		t.Fatalf("ScanJavaScriptWithV8 returned error: %v", err)
	}

	hit := findDetectedByName(result, "Google Analytics Tracking")
	if hit == nil {
		t.Fatalf("expected Google Analytics Tracking to be detected")
	}

	if hit.Count != 1 {
		t.Fatalf("expected deduplicated count to be 1, got %d", hit.Count)
	}
}

func TestScanJavaScriptWithV8_RespectsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ScanJavaScriptWithV8(ctx, "<html><script>navigator.gpu</script></html>")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
