package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"doligo_001/internal/api/handlers"
	apiMiddleware "doligo_001/internal/api/middleware"
	"doligo_001/internal/api/validator"
	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/logger"
	"doligo_001/internal/infrastructure/pdf"
	"doligo_001/internal/infrastructure/repository"
	"doligo_001/internal/usecase/auth"
	bom_uc "doligo_001/internal/usecase/bom"
	invoice_uc "doligo_001/internal/usecase/invoice"
	item_uc "doligo_001/internal/usecase/item"
	margin_uc "doligo_001/internal/usecase/margin"
	stock_uc "doligo_001/internal/usecase/stock"
	thirdparty_uc "doligo_001/internal/usecase/thirdparty"
)

// initServices initializes database-dependent services and returns the db connection
func initServices(ctx context.Context, cfg *config.Config, e *echo.Echo) (*gorm.DB, error) {
	log.Info().Msg("Starting database and services initialization...")

	gormDB, dsn, err := db.InitDatabase(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Database connection established.")

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	if err := db.RunMigrations(ctx, sqlDB, cfg.Database.Type, dsn); err != nil {
		log.Error().Err(err).Msg("Failed to run database migrations.")
		// Depending on the policy, you might want to return an error here
	} else {
		log.Info().Msg("Database migrations completed successfully.")
	}

	pdfGenerator := pdf.NewMarotoGenerator()

	// Repositories
	userRepo := repository.NewGormUserRepository(gormDB)
	thirdPartyRepo := repository.NewGormThirdPartyRepository(gormDB)
	itemRepo := repository.NewGormItemRepository(gormDB)
	bomRepo := repository.NewGormBomRepository(gormDB)
	marginRepo := repository.NewGormMarginRepository(gormDB)
	invoiceRepo := repository.NewInvoiceRepository(gormDB)
	txManager := db.NewGormTransactioner(gormDB)

	// Usecases
	authUsecase := auth.NewAuthUsecase(userRepo, []byte(cfg.JWT.JWTSecret), time.Hour*24)
	thirdPartyUsecase := thirdparty_uc.NewUsecase(thirdPartyRepo)
	itemUsecase := item_uc.NewUsecase(itemRepo)
	stockUsecase := stock_uc.NewUseCase(gormDB, txManager)
	bomUsecase := bom_uc.NewBOMUsecase(bomRepo)
	marginUsecase := margin_uc.NewMarginUsecase(marginRepo)
	invoiceUsecase := invoice_uc.NewUsecase(invoiceRepo, itemRepo, pdfGenerator)

	// Handlers
	authHandler := handlers.NewAuthHandler(authUsecase)
	thirdPartyHandler := handlers.NewThirdPartyHandler(thirdPartyUsecase)
	itemHandler := handlers.NewItemHandler(itemUsecase)
	stockHandler := handlers.NewStockHandler(stockUsecase)
	bomHandler := handlers.NewBOMHandler(bomUsecase, validator.NewValidator())
	marginHandler := handlers.NewMarginHandler(marginUsecase)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceUsecase)

	// Register routes
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.POST("/login", authHandler.Login)

	v1 := e.Group("/api/v1")
	jwtMiddleware := &apiMiddleware.JWTConfig{Secret: []byte(cfg.JWT.JWTSecret)}
	v1.Use(jwtMiddleware.JWT)

	thirdpartiesGroup := v1.Group("/thirdparties")
	thirdpartiesGroup.POST("", thirdPartyHandler.Create)
	thirdpartiesGroup.GET("", thirdPartyHandler.List)

	itemsGroup := v1.Group("/items")
	itemsGroup.POST("", itemHandler.Create)
	itemsGroup.GET("", itemHandler.List)

	v1.POST("/stock/movements", stockHandler.CreateStockMovement)

	bomGroup := v1.Group("/boms")
	bomGroup.POST("", bomHandler.CreateBOM)
	bomGroup.GET("/:id", bomHandler.GetBOMByID)
	bomGroup.GET("/product/:productID", bomHandler.GetBOMByProductID)
	bomGroup.GET("", bomHandler.ListBOMs)
	bomGroup.PUT("/:id", bomHandler.UpdateBOM)
	bomGroup.DELETE("/:id", bomHandler.DeleteBOM)
	bomGroup.POST("/calculate-cost", bomHandler.CalculatePredictiveCost)
	bomGroup.POST("/produce", bomHandler.ProduceItem)

	marginGroup := v1.Group("/margin")
	marginGroup.GET("/products/:productID", marginHandler.GetProductMarginReport)
	marginGroup.GET("", marginHandler.ListOverallMarginReports)

	invoiceGroup := v1.Group("/invoices")
	invoiceGroup.POST("", invoiceHandler.CreateInvoice)
	invoiceGroup.GET("/:id", invoiceHandler.GetInvoice)
	invoiceGroup.GET("/:id/pdf", invoiceHandler.GenerateInvoicePDF)

	log.Info().Msg("All services initialized and routes registered.")
	return gormDB, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading configuration")
	}
	logger.InitLogger(cfg.Log.Level)

	log.Info().Msgf("Application Environment: %s", cfg.AppEnv)

	// Create a root context that is canceled on OS signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	e := echo.New()
	e.Validator = validator.NewValidator()
	e.Use(middleware.Recover())

	gormDB, err := initServices(ctx, cfg, e)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize services")
	}

	// Start server in a goroutine
	go func() {
		log.Info().Msgf("Starting server on port %s", cfg.Port)
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("HTTP server failed to start")
			stop() // trigger shutdown
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	log.Info().Msg("Shutting down server...")

	// Create a context with a timeout for the shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Perform graceful shutdown for the HTTP server
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close database connection
	if gormDB != nil {
		sqlDB, err := gormDB.DB()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get DB instance for closing")
		} else {
			log.Info().Msg("Closing database connection...")
			if err := sqlDB.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close database connection")
			}
		}
	}

	log.Info().Msg("Server gracefully stopped.")
}