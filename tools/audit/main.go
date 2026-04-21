package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	url         = "http://localhost:3000/next-id"
	totalReqs   = 200000 // Aumentado para 200k para teste massivo
	concurrency = 500    // Aumentado para 500 conexões simultâneas
)

type response struct {
	ID uint64 `json:"id"`
}

func main() {
	fmt.Printf("🔥 INICIANDO AUDITORIA BRUTAL: %d requisições | %d workers\n", totalReqs, concurrency)

	var wg sync.WaitGroup
	results := make(chan uint64, totalReqs)
	
	// Usamos um cliente customizado com pool de conexões otimizado
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        concurrency,
			MaxIdleConnsPerHost: concurrency,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	
	// 1. Disparar os trabalhadores (Workers)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Cada worker faz uma fatia do total de requisições
			reqsPerWorker := totalReqs / concurrency
			for j := 0; j < reqsPerWorker; j++ {
				resp, err := client.Get(url)
				if err != nil {
					// Em testes extremos, alguns erros de rede podem ocorrer (ex: portas esgotadas)
					// Mas os IDs recebidos DEVEM ser únicos.
					continue
				}
				
				var r response
				if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
					resp.Body.Close()
					continue
				}
				resp.Body.Close()
				
				results <- r.ID
			}
		}()
	}

	// 2. Fechar canal após conclusão dos workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// 3. Verificação de Duplicatas em Tempo Real
	fmt.Println("⏳ Coletando e verificando IDs...")
	receivedIDs := make(map[uint64]bool)
	duplicates := 0
	totalReceived := 0

	for id := range results {
		totalReceived++
		if _, exists := receivedIDs[id]; exists {
			fmt.Printf("🚨 DUPLICATA ENCONTRADA: %d\n", id)
			duplicates++
		}
		receivedIDs[id] = true
		
		// Feedback visual a cada 50k
		if totalReceived%50000 == 0 {
			fmt.Printf("... %d IDs analisados\n", totalReceived)
		}
	}

	duration := time.Since(start)

	// 4. Relatório Final
	fmt.Println("\n--- 🏁 RELATÓRIO DE AUDITORIA BRUTAL ---")
	fmt.Printf("Duração:          %v\n", duration)
	fmt.Printf("Requisições/seg:  %.2f\n", float64(totalReceived)/duration.Seconds())
	fmt.Printf("IDs Coletados:    %d\n", totalReceived)
	fmt.Printf("Duplicatas:       %d\n", duplicates)
	
	if duplicates == 0 && totalReceived > (totalReqs*95/100) {
		fmt.Println("✅ SUCESSO ABSOLUTO: Nenhuma duplicata em condições de estresse extremo!")
	} else if duplicates > 0 {
		fmt.Println("❌ FALHA CRÍTICA: O sistema gerou IDs duplicados sob carga.")
	} else {
		fmt.Println("⚠️ AVISO: O teste terminou com muitas falhas de rede, mas sem duplicatas.")
	}
}
