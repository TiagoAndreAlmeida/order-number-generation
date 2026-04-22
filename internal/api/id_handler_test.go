package api

import (
	"encoding/json"
	"gopher-id/internal/repository"
	"gopher-id/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIDHandler_GetNextID(t *testing.T) {
	// Setup das dependências
	tmpDir := t.TempDir()
	db, _ := repository.NewDB(tmpDir)
	defer db.Close()
	
	svc, _ := service.NewIDService(db, "test-api", 100)
	defer svc.Close()

	handler := NewIDHandler(svc)

	t.Run("Sucesso - Retorna ID 0 no primeiro request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/next-id", nil)
		w := httptest.NewRecorder()

		handler.GetNextID(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("esperava status 200, obteve %d", w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("esperava content-type application/json, obteve %s", w.Header().Get("Content-Type"))
		}

		var resp NextIDResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("falha ao decodificar JSON: %v", err)
		}

		if resp.ID != 0 {
			t.Errorf("esperava ID 0, obteve %d", resp.ID)
		}
	})

	t.Run("Sucesso - IDs incrementais", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/next-id", nil)
		w := httptest.NewRecorder()

		handler.GetNextID(w, req)

		var resp NextIDResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.ID != 1 {
			t.Errorf("esperava ID 1 na segunda chamada, obteve %d", resp.ID)
		}
	})

	t.Run("Erro - Falha no Serviço", func(t *testing.T) {
		// Criamos um handler com o serviço já fechado para forçar erro
		svc.Close() 
		
		req := httptest.NewRequest("GET", "/next-id", nil)
		w := httptest.NewRecorder()

		handler.GetNextID(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("esperava status 500 ao falhar o serviço, obteve %d", w.Code)
		}
	})
}
