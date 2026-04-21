package service

import (
	"fmt"
	"log"

	"gopher-id/internal/repository"

	"github.com/dgraph-io/badger/v4"
)

// IDService gerencia a lógica de geração de identificadores únicos
type IDService struct {
	seq *badger.Sequence
}

// NewIDService inicializa o serviço de IDs criando uma sequência no repositório.
// A 'key' identifica qual sequência estamos usando (ex: "order_ids") e
// 'bandwidth' define quantos IDs serão reservados na memória por vez.
func NewIDService(repo *repository.DB, key string, bandwidth uint64) (*IDService, error) {
	// 1. Solicita ao repositório a inicialização da sequência no BadgerDB
	seq, err := repo.GetSequence(key, bandwidth)
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar sequência no repositório: %w", err)
	}

	log.Printf("IDService inicializado para a chave '%s' com banda de %d", key, bandwidth)
	return &IDService{seq: seq}, nil
}

// GetNextID obtém o próximo identificador único da sequência.
// Esta operação é thread-safe e extremamente rápida, pois utiliza sync/atomic internamente.
func (s *IDService) GetNextID() (uint64, error) {
	id, err := s.seq.Next()
	if err != nil {
		return 0, fmt.Errorf("erro ao gerar próximo ID: %w", err)
	}
	return id, nil
}

// Close libera a sequência reservada na memória, garantindo que o ponto de controle
// seja atualizado no disco antes do desligamento do servidor.
func (s *IDService) Close() error {
	if s.seq != nil {
		log.Println("Liberando sequência do IDService...")
		return s.seq.Release()
	}
	return nil
}
