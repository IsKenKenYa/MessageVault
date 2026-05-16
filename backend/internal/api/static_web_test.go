package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IsKenKenYa/Commory/backend/internal/config"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func TestServerServesWebRootAndKeepsAPI(t *testing.T) {
	handler := newTestWebHandler(t)

	setupReq := httptest.NewRequest(http.MethodGet, "/api/setup", nil)
	setupRes := httptest.NewRecorder()
	handler.ServeHTTP(setupRes, setupReq)
	if setupRes.Code != http.StatusOK {
		t.Fatalf("setup status = %d, body = %s", setupRes.Code, setupRes.Body.String())
	}
	if strings.Contains(setupRes.Body.String(), "<div id=\"app\">") {
		t.Fatal("api response unexpectedly returned web index")
	}

	indexReq := httptest.NewRequest(http.MethodGet, "/", nil)
	indexRes := httptest.NewRecorder()
	handler.ServeHTTP(indexRes, indexReq)
	if indexRes.Code != http.StatusOK {
		t.Fatalf("index status = %d, body = %s", indexRes.Code, indexRes.Body.String())
	}
	if !strings.Contains(indexRes.Body.String(), "<div id=\"app\">") {
		t.Fatal("expected web index body")
	}
}

func TestServerServesAssetsAndFallsBackToSPA(t *testing.T) {
	handler := newTestWebHandler(t)

	assetReq := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	assetRes := httptest.NewRecorder()
	handler.ServeHTTP(assetRes, assetReq)
	if assetRes.Code != http.StatusOK {
		t.Fatalf("asset status = %d, body = %s", assetRes.Code, assetRes.Body.String())
	}
	if !strings.Contains(assetRes.Body.String(), "commory asset") {
		t.Fatal("expected asset body")
	}

	routeReq := httptest.NewRequest(http.MethodGet, "/dashboard/console", nil)
	routeRes := httptest.NewRecorder()
	handler.ServeHTTP(routeRes, routeReq)
	if routeRes.Code != http.StatusOK {
		t.Fatalf("spa route status = %d, body = %s", routeRes.Code, routeRes.Body.String())
	}
	if !strings.Contains(routeRes.Body.String(), "<div id=\"app\">") {
		t.Fatal("expected spa fallback index body")
	}
}

func TestAPINotFoundDoesNotFallbackToWeb(t *testing.T) {
	handler := newTestWebHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/not-found", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code == http.StatusOK {
		t.Fatalf("expected api not-found/auth failure, got %d with body %s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "<div id=\"app\">") {
		t.Fatal("api route unexpectedly fell back to web index")
	}
}

func newTestWebHandler(t *testing.T) http.Handler {
	t.Helper()
	webRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(webRoot, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webRoot, "index.html"), []byte(`<!doctype html><div id="app"></div>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webRoot, "assets", "app.js"), []byte(`console.log("commory asset")`), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := storage.NewSQLiteProvider(filepath.Join(t.TempDir(), "commory.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	validator, err := msglayer.NewValidator(repoPath("msglayer", "schema", "v0.1", "root.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(config.Config{AuthSecret: "test-secret", TLS: false, WebRoot: webRoot}, store, validator)
	return server.Handler()
}
