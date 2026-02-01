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
		"doligo_001/internal/api"
		"doligo_001/internal/api/handlers"
		apiMiddleware "doligo_001/internal/api/middleware"
		"doligo_001/internal/usecase/auth"
		item_uc "doligo_001/internal/usecase/item"
		thirdparty_uc "doligo_001/internal/usecase/thirdparty"
		stock_uc "doligo_001/internal/usecase/stock"
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
	
		// Initialize use cases
		authUsecase := auth.NewAuthUsecase(userRepo, []byte(cfg.JWT.JWTSecret), time.Hour*24)
		thirdPartyUsecase := thirdparty_uc.NewUsecase(thirdPartyRepo)
		itemUsecase := item_uc.NewUsecase(itemRepo)

		// Initialize transaction manager
		txManager := db.NewGormTransactioner(gormDB)

		// Initialize stock use case
		stockUsecase := stock_uc.NewUseCase(gormDB, txManager)
	
		// Initialize handlers
		authHandler := handlers.NewAuthHandler(authUsecase)
		thirdPartyHandler := handlers.NewThirdPartyHandler(thirdPartyUsecase)
		itemHandler := handlers.NewItemHandler(itemUsecase)
		stockHandler := handlers.NewStockHandler(stockUsecase)
	
		// Setup Echo
		e := echo.New()
		e.Validator = api.NewValidator()
		e.Use(middleware.Recover()) // Echo Recovery Middleware
	
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
			stockHandler.RegisterRoutes(v1.Group("")) // stock handler registers sub-groups
	
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
		log.Info().Msg("Shutting down server...")
	
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 seconds for graceful shutdown
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Server forced to shutdown")
		}
	
		log.Info().Msg("Server gracefully stopped.")
	}
	