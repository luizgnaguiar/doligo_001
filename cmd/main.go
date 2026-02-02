package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"

	"doligo_001/internal/api/handlers"
	apiMiddleware "doligo_001/internal/api/middleware"
	"doligo_001/internal/api/validator"
	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/logger"
	"doligo_001/internal/infrastructure/metrics"
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
func initServices(ctx context.Context, cfg *config.Config, e *echo.Echo) (*gorm.DB, *metrics.Metrics, error) {
	slog.Info("Starting database and services initialization...")

	// Metrics service
	appMetrics := metrics.NewMetrics()

	gormDB, dsn, err := db.InitDatabase(ctx, &cfg.Database)
	if err != nil {
		return nil, nil, err
	}
	slog.Info("Database connection established.")

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, nil, err
	}

	if err := db.RunMigrations(ctx, sqlDB, cfg.Database.Type, dsn); err != nil {
		slog.Error("Failed to run database migrations.", "error", err)
		// Depending on the policy, you might want to return an error here
	} else {
		slog.Info("Database migrations completed successfully.")
	}

	pdfGenerator := pdf.NewMarotoGenerator()
	txManager := db.NewGormTransactioner(gormDB)

	// Repositories
	userRepo := repository.NewGormUserRepository(gormDB)
	thirdPartyRepo := repository.NewGormThirdPartyRepository(gormDB)
	itemRepo := repository.NewGormItemRepository(gormDB)
	bomRepo := repository.NewGormBomRepository(gormDB, txManager)
	marginRepo := repository.NewGormMarginRepository(gormDB)
	invoiceRepo := repository.NewInvoiceRepository(gormDB)
	stockRepo := repository.NewGormStockRepository(gormDB)
	stockMoveRepo := repository.NewGormStockMovementRepository(gormDB)
	stockLedgerRepo := repository.NewGormStockLedgerRepository(gormDB)
	warehouseRepo := repository.NewGormWarehouseRepository(gormDB)
	binRepo := repository.NewGormBinRepository(gormDB)

	// Usecases
	authUsecase := auth.NewAuthUsecase(userRepo, []byte(cfg.JWT.JWTSecret), time.Hour*24)
	thirdPartyUsecase := thirdparty_uc.NewUsecase(thirdPartyRepo)
	itemUsecase := item_uc.NewUsecase(itemRepo)
	stockUsecase := stock_uc.NewUseCase(txManager, stockRepo, stockMoveRepo, stockLedgerRepo, warehouseRepo, binRepo, itemRepo)
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
	metricsHandler := handlers.NewMetricsHandler(appMetrics)

	// Register routes
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.GET("/ready", func(c echo.Context) error {
		if err := db.Ping(c.Request().Context(), gormDB); err != nil {
			slog.Error("Readiness check failed: database ping failed", "error", err)
			return c.String(http.StatusServiceUnavailable, "Database not ready")
		}
		return c.String(http.StatusOK, "Ready")
	})
	e.GET("/metrics/internal", metricsHandler.GetMetrics)

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

	slog.Info("All services initialized and routes registered.")
	return gormDB, appMetrics, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}
	logger.InitLogger(cfg.Log.Level)

	slog.Info("Application Environment", "env", cfg.AppEnv)

	// Create a root context that is canceled on OS signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	e := echo.New()
	e.Validator = validator.NewValidator()
	e.Use(middleware.Recover())

	gormDB, appMetrics, err := initServices(ctx, cfg, e)
	if err != nil {
		slog.Error("Failed to initialize services", "error", err)
		os.Exit(1)
	}

	// Setup middlewares
	e.Use(apiMiddleware.RequestLogger)
	e.Use(apiMiddleware.MetricsMiddleware(appMetrics))

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed to start", "error", err)
			stop() // trigger shutdown
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	slog.Info("Shutting down server...")

	// Create a context with a timeout for the shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Perform graceful shutdown for the HTTP server
	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	// Close database connection
	if gormDB != nil {
		sqlDB, errSql := gormDB.DB()
		if errSql != nil {
			slog.Error("Failed to get DB instance for closing", "error", errSql)
		} else {
			slog.Info("Closing database connection...")
			if errDbClose := sqlDB.Close(); errDbClose != nil {
				slog.Error("Failed to close database connection", "error", errDbClose)
			}
		}
	}

	slog.Info("Server gracefully stopped.")
}