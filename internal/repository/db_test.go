package repository

import (
	"testing"
)

// TestNewDB_Success valida se o banco pode ser aberto e fechado corretamente
func TestNewDB_Success(t *testing.T) {
	// t.TempDir() cria automaticamente uma pasta limpa para o teste
	tmpDir := t.TempDir()

	db, err := NewDB(tmpDir)
	if err != nil {
		t.Fatalf("falha esperada ao abrir banco: %v", err)
	}

	if db.Conn == nil {
		t.Fatal("conexão com banco não deveria ser nula")
	}

	// Testa o fechamento
	if err := db.Close(); err != nil {
		t.Errorf("falha ao fechar banco: %v", err)
	}
}

// TestNewDB_PermissionError valida erro ao tentar abrir banco em local proibido
func TestNewDB_PermissionError(t *testing.T) {
	// Caminho que normalmente exige root no Linux
	path := "/proc/test_gopher_id"

	_, err := NewDB(path)
	if err == nil {
		t.Error("esperava erro ao tentar criar diretório em local proibido, mas obteve nil")
	}
}

// TestNewDB_Locking valida que o BadgerDB impede duas conexões no mesmo diretório
func TestNewDB_Locking(t *testing.T) {
	tmpDir := t.TempDir()

	// Abre a primeira conexão
	db1, err := NewDB(tmpDir)
	if err != nil {
		t.Fatalf("falha ao abrir primeira conexão: %v", err)
	}
	defer db1.Close()

	// Tenta abrir a segunda conexão no MESMO diretório
	_, err = NewDB(tmpDir)
	if err == nil {
		t.Error("esperava erro de 'database locked' ao tentar abrir segunda conexão no mesmo diretório, mas obteve nil")
	}
}

// TestGetSequence valida a inicialização de uma sequência
func TestGetSequence(t *testing.T) {
	tmpDir := t.TempDir()
	db, _ := NewDB(tmpDir)
	defer db.Close()

	seq, err := db.GetSequence("test-key", 100)
	if err != nil {
		t.Fatalf("falha ao criar sequência: %v", err)
	}

	if seq == nil {
		t.Fatal("sequência não deveria ser nula")
	}
	defer seq.Release()
}
