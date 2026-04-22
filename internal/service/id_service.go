package service

import (
	"fmt"
	"log"
	"sync"

	"gopher-id/internal/repository"

	"github.com/dgraph-io/badger/v4"
)

// IDService gerencia a lógica de geração de identificadores únicos
type IDService struct {
	seq    *badger.Sequence
	closed bool
	mu     sync.RWMutex
}

// NewIDService inicializa o serviço de IDs criando uma sequência no repositório.
func NewIDService(repo *repository.DB, key string, bandwidth uint64) (*IDService, error) {
	seq, err := repo.GetSequence(key, bandwidth)
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar sequência no repositório: %w", err)
	}

	log.Printf("IDService inicializado para a chave '%s' com banda de %d", key, bandwidth)
	return &IDService{seq: seq}, nil
}

// GetNextID obtém o próximo identificador único da sequência.
func (s *IDService) GetNextID() (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return 0, fmt.Errorf("serviço de IDs está desativado")
	}

	id, err := s.seq.Next()
	if err != nil {
		return 0, fmt.Errorf("erro ao gerar próximo ID: %w", err)
	}
	return id, nil
}

// Close libera a sequência reservada na memória.
func (s *IDService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	if s.seq != nil {
		log.Println("Liberando sequência do IDService...")
		return s.seq.Release()
	}
	return nil
}
