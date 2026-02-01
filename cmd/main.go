package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	stdlog "log" // Alias standard log to avoid conflicts with zerolog.log

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/logger"
	"doligo_001/internal/infrastructure/repository"
	"doligo_001/internal/api/validator"
	"doligo_001/internal/api/handlers"
	apiMiddleware "doligo_001/internal/api/middleware"
	"doligo_001/internal/usecase/auth"
	item_uc "doligo_001/internal/usecase/item"
	thirdparty_uc "doligo_001/internal/usecase/thirdparty"
	stock_uc "doligo_001/internal/usecase/stock"
	bom_uc "doligo_001/internal/usecase/bom"
	margin_uc "doligo_001/internal/usecase/margin" // Import margin usecase
	"doligo_001/internal/infrastructure/worker" // Import worker package
)

func main() {
	// Initialize logger first
	cfg, err := config.LoadConfig()
	if err != nil {
		stdlog.Fatalf("Error loading configuration: %v", err)
	}
	logger.InitLogger(cfg.Log.Level)

	log.Info().Msgf("Application Environment: %s", cfg.AppEnv)
	log.Info().Msgf("Listening on Port: %s", cfg.Port)
	log.Info().Msgf("Database Type: %s", cfg.Database.Type)
	log.Info().Msgf("Log Level: %s", cfg.Log.Level)

	// Initialize database
	gormDB, dsn, err := db.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing database")
	}
	defer func() {
		sqlDB, closeErr := gormDB.DB()
		if closeErr == nil {
			closeErr = sqlDB.Close()
		}
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close database connection")
		} else {
			log.Info().Msg("Database connection closed")
		}
	}()

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting generic database object from GORM")
	}

	// Run migrations
	if err := db.RunMigrations(sqlDB, cfg.Database.Type, dsn); err != nil {
		log.Fatal().Err(err).Msg("Error running database migrations")
	}

	// Initialize repositories
	userRepo := repository.NewGormUserRepository(gormDB)
	thirdPartyRepo := repository.NewGormThirdPartyRepository(gormDB)
	itemRepo := repository.NewGormItemRepository(gormDB)
	bomRepo := repository.NewGormBomRepository(gormDB)
	marginRepo := repository.NewGormMarginRepository(gormDB) // Initialize margin repository

	// Initialize transaction manager
	txManager := db.NewGormTransactioner(gormDB)

	// Initialize stock repositories


	// Initialize Worker Pool for PDF generation
	// These values (workers, buffer) should ideally come from configuration.
	pdfWorkerPool := worker.NewWorkerPool(5, 10, "PDFGenerator")
	log.Info().Msg("PDF Worker Pool initialized.")

	// Initialize use cases
	authUsecase := auth.NewAuthUsecase(userRepo, []byte(cfg.JWT.JWTSecret), time.Hour*24)
	thirdPartyUsecase := thirdparty_uc.NewUsecase(thirdPartyRepo)
	itemUsecase := item_uc.NewUsecase(itemRepo)
	stockUsecase := stock_uc.NewUseCase(gormDB, txManager)
	bomUsecase := bom_uc.NewBOMUsecase(bomRepo)
	marginUsecase := margin_uc.NewMarginUsecase(marginRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authUsecase)
	thirdPartyHandler := handlers.NewThirdPartyHandler(thirdPartyUsecase)
	itemHandler := handlers.NewItemHandler(itemUsecase)
	stockHandler := handlers.NewStockHandler(stockUsecase)
	bomHandler := handlers.NewBOMHandler(bomUsecase, validator.NewValidator())
	marginHandler := handlers.NewMarginHandler(marginUsecase) // Initialize margin handler

	// Setup Echo
	e := echo.New()
	e.Validator = validator.NewValidator()
	e.Use(middleware.Recover())

	// Register public routes
	authHandler.RegisterRoutes(e)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// API v1 Group with JWT Middleware
	v1 := e.Group("/api/v1")
	jwtMiddleware := &apiMiddleware.JWTConfig{
		Secret: []byte(cfg.JWT.JWTSecret),
	}
	v1.Use(jwtMiddleware.JWT)

	// Register authenticated routes
	thirdPartyHandler.RegisterRoutes(v1.Group("/thirdparties"))
	itemHandler.RegisterRoutes(v1.Group("/items"))
	stockHandler.RegisterRoutes(v1.Group(""))

	// Register BOM routes
	bomGroup := v1.Group("/boms")
	bomGroup.POST("", bomHandler.CreateBOM)
	bomGroup.GET("/:id", bomHandler.GetBOMByID)
	bomGroup.GET("/product/:productID", bomHandler.GetBOMByProductID)
	bomGroup.GET("", bomHandler.ListBOMs)
	bomGroup.PUT("/:id", bomHandler.UpdateBOM)
	bomGroup.DELETE("/:id", bomHandler.DeleteBOM)
	bomGroup.POST("/calculate-cost", bomHandler.CalculatePredictiveCost)
			bomGroup.POST("/produce", bomHandler.ProduceItem)
	
		// Register Margin routes
		marginHandler.RegisterRoutes(v1.Group("/margin"))
	
		// Start server in a goroutine
		go func() {
			if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Shutting down the server")
			}
		}()
	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server and PDF Worker Pool...")

	// Create a context for the overall shutdown with enough time for both server and worker pool
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Max 15s for worker pool
	defer cancel()

	// Shutdown Echo server
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Shutdown Worker Pool
	pdfWorkerPool.Shutdown(15 * time.Second) // CT-04: Task Runner finishes tasks up to 15 seconds

	log.Info().Msg("Server and PDF Worker Pool gracefully stopped.")
}