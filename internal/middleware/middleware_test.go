package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	// 1. Criar um handler que causa um panic proposital
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("erro catastrófico proposital")
	})

	// 2. Envolver o handler com o middleware Recovery
	handler := Recovery(panicHandler)

	// 3. Simular uma requisição
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// 4. Executar e garantir que não houve crash (o teste deve continuar)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("O middleware Recovery falhou em capturar o panic!")
		}
	}()

	handler.ServeHTTP(w, req)

	// 5. Validar se o status retornado é 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperava status 500 após panic, obteve %d", w.Code)
	}
}
