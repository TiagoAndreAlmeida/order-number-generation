package repository

import (
	"log"
	"os"

	"github.com/dgraph-io/badger/v4"
)

// DB representa a conexão com o banco de dados BadgerDB
type DB struct {
	Conn *badger.DB
}

// NewDB inicializa e abre a conexão com o BadgerDB
func NewDB(path string) (*DB, error) {
	// 1. Garantir que o diretório de dados exista
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	// 2. Configurar as opções do BadgerDB
	// DefaultOptions já vem otimizado para performance
	opts := badger.DefaultOptions(path).
		WithLoggingLevel(badger.ERROR) // Evita logs excessivos no console

	// 3. Abrir a conexão
	conn, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	log.Printf("BadgerDB aberto com sucesso em: %s", path)
	return &DB{Conn: conn}, nil
}

// Close fecha a conexão com o banco de dados de forma segura
func (db *DB) Close() error {
	if db.Conn != nil {
		log.Println("Fechando conexão com BadgerDB...")
		return db.Conn.Close()
	}
	return nil
}

// GetSequence inicializa uma sequência no BadgerDB com uma banda (bandwidth) específica.
// A banda define quantos IDs serão reservados na memória para alta performance.
func (db *DB) GetSequence(key string, bandwidth uint64) (*badger.Sequence, error) {
	return db.Conn.GetSequence([]byte(key), bandwidth)
}
