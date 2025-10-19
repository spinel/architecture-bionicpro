package main

import (
	"bionicpro-backend/internal/auth"
	"bionicpro-backend/internal/config"
	"bionicpro-backend/internal/database"
	"bionicpro-backend/internal/handlers"
	"bionicpro-backend/internal/middleware"
	"bionicpro-backend/internal/repository"
	"bionicpro-backend/internal/service"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Инициализация логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация базы данных
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Инициализация репозиториев
	reportRepo := repository.NewReportRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Инициализация сервисов
	var authService auth.AuthService = auth.NewDemoService(&cfg.Auth)
	reportService := service.NewReportService(reportRepo, logger)
	analyticsService := service.NewAnalyticsService(analyticsRepo)

	// Инициализация обработчиков
	reportHandler := handlers.NewReportHandler(reportService, authService, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	healthHandler := handlers.NewHealthHandler(db)

	// Инициализация middleware
	accessControlMiddleware := middleware.NewAccessControlMiddleware(logger)
	demoAccessControlMiddleware := middleware.NewDemoAccessControlMiddleware(logger)

	// Настройка Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Настройка CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Middleware для аутентификации
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Маршруты API
	api := router.Group("/api/v1")
	{
		// Публичные маршруты
		api.GET("/health", healthHandler.Health)
		api.GET("/ready", healthHandler.Ready)

		// Маршруты аналитики (защищённые)
		analytics := api.Group("/analytics")
		analytics.Use(authMiddleware.RequireAuth())
		{
			analytics.GET("/reports", accessControlMiddleware.RequireOwnDataFromQuery(), analyticsHandler.GetUserReport)
			analytics.GET("/summary", analyticsHandler.GetDailySummary) // Сводка доступна всем аутентифицированным
			analytics.GET("/user/:user_id", accessControlMiddleware.RequireOwnData(), analyticsHandler.GetUserReportByID)
		}

		// Демо-маршруты для демонстрации контроля доступа (без аутентификации)
		demo := api.Group("/demo")
		{
			demo.GET("/reports", demoAccessControlMiddleware.RequireOwnDataFromQueryDemo(), analyticsHandler.GetUserReport)
			demo.GET("/user/:user_id", demoAccessControlMiddleware.RequireOwnDataDemo(), analyticsHandler.GetUserReportByID)
		}

		// Защищенные маршруты
		protected := api.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Маршруты отчетов (старые)
			reports := protected.Group("/reports")
			{
				reports.GET("", reportHandler.GetReports)
				reports.GET("/:id", reportHandler.GetReport)
				reports.POST("/generate", reportHandler.GenerateReport)
				reports.GET("/download/:id", reportHandler.DownloadReport)
			}
		}
	}

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("Сервер запущен на порту %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
