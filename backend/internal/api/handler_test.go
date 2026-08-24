package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vulnkit/internal/api"
	"vulnkit/internal/docker"
	"vulnkit/internal/presets"
	"vulnkit/tests"
)

func newTestHandler(t *testing.T) *api.Handler {
	t.Helper()
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	dockerClient := &docker.Client{}
	return api.NewHandler(dockerClient, store, t.TempDir())
}

func TestHealth_Returns200(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
}

func TestGenerateCompose_ValidRequest(t *testing.T) {
	h := newTestHandler(t)
	body := `{
		"name": "test-lab",
		"services": [
			{"name":"mysql","image":"mysql","version":"5.6.51","ports":["3306:3306"],"env_vars":{}}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/labs/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.GenerateCompose(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !strings.Contains(resp["yaml"], "mysql:5.6.51") {
		t.Errorf("expected yaml to contain mysql:5.6.51, got: %s", resp["yaml"])
	}
}

func TestGenerateCompose_EmptyServices(t *testing.T) {
	h := newTestHandler(t)
	body := `{"name":"test-lab","services":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/labs/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.GenerateCompose(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGenerateCompose_InvalidJSON(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/labs/generate", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.GenerateCompose(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListPresets_EmptyInitially(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/presets", nil)
	w := httptest.NewRecorder()

	h.ListPresets(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	presetList, ok := resp["presets"].([]any)
	if !ok {
		t.Fatalf("expected presets array, got %T", resp["presets"])
	}
	if len(presetList) != 0 {
		t.Errorf("expected empty list, got %d items", len(presetList))
	}
}

func TestSavePreset_ValidRequest(t *testing.T) {
	h := newTestHandler(t)
	body := `{
		"name": "My SQLi Lab",
		"tags": ["SQLi"],
		"services": [
			{"name":"mysql","image":"mysql","version":"5.6.51","ports":["3306:3306"],"env_vars":{}}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/presets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SavePreset(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var saved map[string]any
	json.Unmarshal(w.Body.Bytes(), &saved)
	if saved["id"] == "" || saved["id"] == nil {
		t.Error("expected saved preset to have an ID")
	}
	if saved["name"] != "My SQLi Lab" {
		t.Errorf("expected name 'My SQLi Lab', got %v", saved["name"])
	}
}

func TestSaveAndListPreset_RoundTrip(t *testing.T) {
	h := newTestHandler(t)

	saveBody := `{"name":"Round trip","tags":["XSS"],"services":[]}`
	saveReq := httptest.NewRequest(http.MethodPost, "/api/presets", strings.NewReader(saveBody))
	saveReq.Header.Set("Content-Type", "application/json")
	saveW := httptest.NewRecorder()
	h.SavePreset(saveW, saveReq)

	if saveW.Code != http.StatusCreated {
		t.Fatalf("save failed: %d %s", saveW.Code, saveW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/presets", nil)
	listW := httptest.NewRecorder()
	h.ListPresets(listW, listReq)

	var resp map[string]any
	json.Unmarshal(listW.Body.Bytes(), &resp)
	items := resp["presets"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(items))
	}
	item := items[0].(map[string]any)
	if item["name"] != "Round trip" {
		t.Errorf("expected name 'Round trip', got %v", item["name"])
	}
}

func TestSavePreset_InvalidJSON(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/presets", strings.NewReader("bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SavePreset(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
