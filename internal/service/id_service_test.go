package service

import (
	"gopher-id/internal/repository"
	"sync"
	"testing"
)

// TestIDService_SequentialID valida se os IDs gerados são sequenciais e incrementais
func TestIDService_SequentialID(t *testing.T) {
	tmpDir := t.TempDir()
	db, _ := repository.NewDB(tmpDir)
	defer db.Close()

	svc, err := NewIDService(db, "test-key", 100)
	if err != nil {
		t.Fatalf("falha ao inicializar serviço: %v", err)
	}
	defer svc.Close()

	// Começamos com um valor que permite que o 0 seja o primeiro ID válido
	for i := 0; i < 10; i++ {
		currentID, err := svc.GetNextID()
		if err != nil {
			t.Fatalf("falha ao obter ID no loop %d: %v", i, err)
		}

		// Apenas validamos se o ID é o que esperamos na sequência (0, 1, 2...)
		if currentID != uint64(i) {
			t.Errorf("ID gerado (%d) é diferente do esperado (%d)", currentID, i)
		}
	}
}

// TestIDService_Persistence valida se a sequência continua após reinício do serviço
func TestIDService_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	key := "persistent-key"
	bandwidth := uint64(100)

	// Primeira rodada: Gera 5 IDs (0, 1, 2, 3, 4) e fecha
	{
		db, _ := repository.NewDB(tmpDir)
		svc, _ := NewIDService(db, key, bandwidth)
		
		for i := 0; i < 5; i++ {
			svc.GetNextID()
		}
		
		svc.Close()
		db.Close()
	}

	// Segunda rodada: Reabre na mesma pasta e verifica o ID
	{
		db, _ := repository.NewDB(tmpDir)
		defer db.Close()
		
		svc, _ := NewIDService(db, key, bandwidth)
		defer svc.Close()

		nextID, err := svc.GetNextID()
		if err != nil {
			t.Fatalf("falha ao obter próximo ID após reinício: %v", err)
		}

		// O próximo ID deve ser 5 ou maior (caso o Badger pule o range por segurança)
		if nextID < 5 {
			t.Errorf("ID após reinício (%d) deveria ser pelo menos 5", nextID)
		}
	}
}

// TestIDService_Concurrency valida se o serviço é seguro sob alta concorrência
func TestIDService_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	db, _ := repository.NewDB(tmpDir)
	defer db.Close()

	// Banda de 1000 para evitar muitas idas ao disco durante o teste
	svc, _ := NewIDService(db, "race-key", 1000)
	defer svc.Close()

	var wg sync.WaitGroup
	ids := sync.Map{} // Mapa seguro para concorrência para detectar duplicatas
	
	workers := 100
	reqsPerWorker := 50
	totalExpected := workers * reqsPerWorker

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < reqsPerWorker; j++ {
				id, err := svc.GetNextID()
				if err != nil {
					t.Errorf("erro em worker concorrente: %v", err)
					return
				}

				// Se o ID já existir no mapa, temos uma falha de duplicata!
				if _, loaded := ids.LoadOrStore(id, true); loaded {
					t.Errorf("🚨 DUPLICATA DETECTADA: ID %d", id)
				}
			}
		}()
	}

	wg.Wait()

	// Verifica se coletamos a quantidade correta de IDs únicos
	count := 0
	ids.Range(func(key, value any) bool {
		count++
		return true
	})

	if count != totalExpected {
		t.Errorf("esperava %d IDs únicos, mas obteve %d", totalExpected, count)
	}
}
