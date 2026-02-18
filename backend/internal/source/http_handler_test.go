package source

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSourceHandlers_DeleteRestoreAndFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewService(NewInMemoryRepository())
	r := gin.New()
	RegisterAdminSourceRoutes(r, svc)

	createBody := []byte(`{
		"name":"UFC Events",
		"source_type":"schedule",
		"platform":"ufc",
		"source_url":"https://www.ufc.com/events",
		"parser_kind":"ufc_schedule",
		"enabled":true,
		"rights_display":true,
		"rights_playback":false,
		"rights_ai_summary":true
	}`)
	createResp := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/admin/sources", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusOK {
		t.Fatalf("expected create 200, got %d", createResp.Code)
	}

	deleteResp := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodDelete, "/admin/sources/1", nil)
	r.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("expected delete 200, got %d", deleteResp.Code)
	}

	listResp := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/admin/sources", nil)
	r.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", listResp.Code)
	}
	items := decodeItems(t, listResp.Body.Bytes())
	if len(items) != 0 {
		t.Fatalf("expected empty active list after delete, got %d", len(items))
	}

	deletedResp := httptest.NewRecorder()
	deletedReq := httptest.NewRequest(http.MethodGet, "/admin/sources?include_deleted=true", nil)
	r.ServeHTTP(deletedResp, deletedReq)
	if deletedResp.Code != http.StatusOK {
		t.Fatalf("expected include_deleted list 200, got %d", deletedResp.Code)
	}
	deletedItems := decodeItems(t, deletedResp.Body.Bytes())
	if len(deletedItems) != 1 {
		t.Fatalf("expected 1 deleted item, got %d", len(deletedItems))
	}

	restoreResp := httptest.NewRecorder()
	restoreReq := httptest.NewRequest(http.MethodPost, "/admin/sources/1/restore", nil)
	r.ServeHTTP(restoreResp, restoreReq)
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("expected restore 200, got %d", restoreResp.Code)
	}

	getResp := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/admin/sources/1", nil)
	r.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected get 200 after restore, got %d", getResp.Code)
	}
}

func decodeItems(t *testing.T, payload []byte) []map[string]any {
	t.Helper()

	var result struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	return result.Items
}
