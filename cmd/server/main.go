package main

import (
	"context"
	"fmt"
	"gopher-id/internal/api"
	"gopher-id/internal/middleware"
	"gopher-id/internal/repository"
	"gopher-id/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Configurações Iniciais
	dbPath := "./data/id_store"
	sequenceKey := "order_ids"
	var bandwidth uint64 = 1000 // Reserva 1000 IDs na RAM por vez

	// 2. Inicialização da Camada de Dados (Repository)
	db, err := repository.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Erro ao abrir banco de dados: %v", err)
	}

	// 3. Inicialização da Camada de Negócio (Service)
	idService, err := service.NewIDService(db, sequenceKey, bandwidth)
	if err != nil {
		db.Close() // Fecha o banco se o serviço falhar
		log.Fatalf("Erro ao inicializar o serviço de IDs: %v", err)
	}

	// 4. Inicialização da Camada de API (Handler)
	idHandler := api.NewIDHandler(idService)

	// 5. Configuração do Roteador Nativo e Rotas
	mux := http.NewServeMux()

	// Rota de Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "up"}`))
	})

	// Rota Principal de Geração de IDs
	mux.HandleFunc("GET /next-id", idHandler.GetNextID)

	// 6. Aplicação dos Middlewares (Encadeamento manual)
	var handler http.Handler = mux
	handler = middleware.Logger(handler)
	handler = middleware.Recovery(handler)

	// 7. Configuração do Servidor HTTP
	srv := &http.Server{
		Addr:         ":3000",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 8. Canal para Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nServidor recebeu sinal de desligamento...")

		// Tempo limite para encerrar requisições pendentes
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Ordem de desligamento: 
		// 1. Parar de aceitar novas requisições HTTP
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Erro no desligamento do servidor HTTP: %v\n", err)
		}

		// 2. Liberar a sequência na memória (commit do range no disco)
		if err := idService.Close(); err != nil {
			log.Printf("Erro ao fechar o serviço de IDs: %v\n", err)
		}

		// 3. Fechar a conexão com o banco de dados BadgerDB
		if err := db.Close(); err != nil {
			log.Printf("Erro ao fechar o banco de dados: %v\n", err)
		}
	}()

	// 9. Inicialização do Servidor
	fmt.Printf("Gopher-ID iniciado na porta %s\n", srv.Addr)
	fmt.Printf("Sequência '%s' ativa com bandwidth de %d\n", sequenceKey, bandwidth)
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar o servidor: %v\n", err)
	}

	fmt.Println("Servidor finalizado com segurança.")
}
