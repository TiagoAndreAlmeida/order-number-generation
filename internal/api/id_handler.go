package api

import (
	"encoding/json"
	"log"
	"net/http"

	"gopher-id/internal/service"
)

// IDHandler gerencia as requisições HTTP relacionadas à geração de IDs
type IDHandler struct {
	service *service.IDService
}

// NewIDHandler cria um novo handler injetando a dependência do IDService
func NewIDHandler(s *service.IDService) *IDHandler {
	return &IDHandler{service: s}
}

// NextIDResponse define a estrutura de resposta JSON para o cliente
type NextIDResponse struct {
	ID uint64 `json:"id"`
}

// GetNextID processa a requisição GET /next-id e retorna um novo identificador único
func (h *IDHandler) GetNextID(w http.ResponseWriter, r *http.Request) {
	// 1. Solicita o próximo ID ao serviço
	id, err := h.service.GetNextID()
	if err != nil {
		log.Printf("Erro ao gerar ID: %v", err)
		http.Error(w, "Erro interno ao gerar identificador", http.StatusInternalServerError)
		return
	}

	// 2. Prepara a resposta em JSON
	resp := NextIDResponse{ID: id}

	// 3. Define os cabeçalhos e envia a resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// O encoder JSON é eficiente para escrever diretamente no ResponseWriter
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
	}
}
