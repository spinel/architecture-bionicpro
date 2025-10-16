package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spinel/architecture-bionicpro/backend/internal/config"
	"github.com/spinel/architecture-bionicpro/backend/internal/handler"
	"github.com/spinel/architecture-bionicpro/backend/internal/middleware"
	"github.com/spinel/architecture-bionicpro/backend/internal/storage"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Reports API service...")
	log.Printf("Server will listen on %s", cfg.Server.GetAddress())

	// Инициализируем подключение к ClickHouse
	clickhouseStorage, err := storage.NewClickHouseStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer clickhouseStorage.Close()
	log.Println("Successfully connected to ClickHouse")

	// Создаём handlers
	reportsHandler := handler.NewReportsHandler(clickhouseStorage)

	// Создаём middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg)
	accessLogger := middleware.NewAccessLogger()

	// Настраиваем роутер
	router := mux.NewRouter()

	// Публичные эндпоинты
	router.HandleFunc("/api/health", reportsHandler.HealthCheck).Methods("GET")

	// Защищённые эндпоинты (требуют аутентификацию)
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Применяем middleware в порядке:
	// 1. Логирование доступа
	// 2. JWT аутентификация
	// 3. Защита от манипуляции user_id
	apiRouter.Use(accessLogger.LogAccess)
	apiRouter.Use(authMiddleware.Authenticate)
	apiRouter.Use(accessLogger.PreventUserIDManipulation)

	apiRouter.HandleFunc("/reports", reportsHandler.GetReports).Methods("GET")
	apiRouter.HandleFunc("/reports/{id}", reportsHandler.GetReportByID).Methods("GET")

	// Настраиваем CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.CORS.AllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Создаём HTTP сервер
	server := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      corsHandler.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server is running on %s", cfg.Server.GetAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
