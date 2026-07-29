package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vistone/fingerprint/modules/crawler"
	gatewaypkg "github.com/vistone/fingerprint/modules/gateway"
)

func TestHandleClientTest_RejectsBlockedSSRFURL(t *testing.T) {
	handler := NewHandler(gatewaypkg.NewGateway(nil))
	if len(handler.profiles) == 0 {
		t.Fatal("expected at least one profile")
	}

	body := []byte(`{"profileId":"` + handler.profiles[0].ID + `","url":"http://127.0.0.1:8080","method":"GET"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/client/test", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.handleClientTest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected blocked target to return 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleCrawlerCrawl_RejectsBlockedSSRFURL(t *testing.T) {
	gw := gatewaypkg.NewGateway(nil)
	gw.SetCrawler(crawler.NewCrawler(nil))
	handler := NewHandler(gw)

	body := []byte(`{"url":"http://127.0.0.1:8080"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/crawler/crawl", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.handleCrawlerCrawl(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected blocked target to return 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
