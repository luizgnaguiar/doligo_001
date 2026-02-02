package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	stdlog "log"

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
	"doligo_001/internal/infrastructure/repository"
	"doligo_001/internal/infrastructure/worker"
	"doligo_001/internal/usecase/auth"
	bom_uc "doligo_001/internal/usecase/bom"
	invoice_uc "doligo_001/internal/usecase/invoice"
	item_uc "doligo_001/internal/usecase/item"
	margin_uc "doligo_001/internal/usecase/margin"
	stock_uc "doligo_001/internal/usecase/stock"
	thirdparty_uc "doligo_001/internal/usecase/thirdparty"
)

// serviceRegistry holds all the application's handlers.
// It's designed to be populated after the database is ready.
type serviceRegistry struct {
	authHandler       *handlers.AuthHandler
	thirdPartyHandler *handlers.ThirdPartyHandler
	itemHandler       *handlers.ItemHandler
	stockHandler      *handlers.StockHandler
	bomHandler        *handlers.BOMHandler
	marginHandler     *handlers.MarginHandler
	invoiceHandler    *handlers.InvoiceHandler
}

// appState holds the service registry and a mutex to control access.
var appState struct {
	sync.RWMutex
	services *serviceRegistry
}

// initServices initializes database-dependent services and populates the appState.
func initServices(cfg *config.Config) {
	log.Info().Msg("Starting database and services initialization...")

	var gormDB *gorm.DB
	gormDB, dsn, err := db.InitDatabase(&cfg.Database)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize database. Service routes will not be available.")
		return
	}
	log.Info().Msg("Database connection established.")

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get generic database object from GORM. Service routes will not be available.")
		return
	}

	if err := db.RunMigrations(sqlDB, cfg.Database.Type, dsn); err != nil {
		log.Error().Err(err).Msg("Failed to run database migrations. Service routes may not work correctly.")
	} else {
		log.Info().Msg("Database migrations completed successfully.")
	}

	// Initialize repositories, usecases, and handlers
	userRepo := repository.NewGormUserRepository(gormDB)
	thirdPartyRepo := repository.NewGormThirdPartyRepository(gormDB)
	itemRepo := repository.NewGormItemRepository(gormDB)
	bomRepo := repository.NewGormBomRepository(gormDB)
	marginRepo := repository.NewGormMarginRepository(gormDB)
	invoiceRepo := repository.NewInvoiceRepository(gormDB)
	txManager := db.NewGormTransactioner(gormDB)
	authUsecase := auth.NewAuthUsecase(userRepo, []byte(cfg.JWT.JWTSecret), time.Hour*24)
	thirdPartyUsecase := thirdparty_uc.NewUsecase(thirdPartyRepo)
	itemUsecase := item_uc.NewUsecase(itemRepo)
	stockUsecase := stock_uc.NewUseCase(gormDB, txManager)
	bomUsecase := bom_uc.NewBOMUsecase(bomRepo)
	marginUsecase := margin_uc.NewMarginUsecase(marginRepo)
	invoiceUsecase := invoice_uc.NewUsecase(invoiceRepo, itemRepo)

	newServices := &serviceRegistry{
		authHandler:       handlers.NewAuthHandler(authUsecase),
		thirdPartyHandler: handlers.NewThirdPartyHandler(thirdPartyUsecase),
		itemHandler:       handlers.NewItemHandler(itemUsecase),
		stockHandler:      handlers.NewStockHandler(stockUsecase),
		bomHandler:        handlers.NewBOMHandler(bomUsecase, validator.NewValidator()),
		marginHandler:     handlers.NewMarginHandler(marginUsecase),
		invoiceHandler:    handlers.NewInvoiceHandler(invoiceUsecase),
	}

	// Atomically update the app state with the new services
	appState.Lock()
	defer appState.Unlock()
	appState.services = newServices

	log.Info().Msg("All database-dependent services initialized and are now available.")
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		stdlog.Fatalf("Error loading configuration: %v", err)
	}
	logger.InitLogger(cfg.Log.Level)

	log.Info().Msgf("Application Environment: %s", cfg.AppEnv)
	log.Info().Msgf("Listening on Port: %s", cfg.Port)

	pdfWorkerPool := worker.NewWorkerPool(5, 10, "PDFGenerator")
	log.Info().Msg("PDF Worker Pool initialized.")

	e := echo.New()
	e.Validator = validator.NewValidator()
	e.Use(middleware.Recover())

	// serviceReady is a wrapper for handlers that depend on DB initialization.
	serviceReady := func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			appState.RLock()
			isReady := appState.services != nil
			appState.RUnlock()

			if !isReady {
				return c.String(http.StatusServiceUnavailable, "Service not ready")
			}
			return h(c)
		}
	}

	// == Register Routes ==
	// Non-dependent routes are registered directly.
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// DB-dependent public routes
	e.POST("/login", serviceReady(func(c echo.Context) error { return appState.services.authHandler.Login(c) }))

	// API v1 Group with JWT Middleware
	v1 := e.Group("/api/v1")
	jwtMiddleware := &apiMiddleware.JWTConfig{Secret: []byte(cfg.JWT.JWTSecret)}
	v1.Use(jwtMiddleware.JWT)

	// Register authenticated routes using the wrapper
	thirdpartiesGroup := v1.Group("/thirdparties")
	thirdpartiesGroup.POST("", serviceReady(func(c echo.Context) error { return appState.services.thirdPartyHandler.Create(c) }))
	thirdpartiesGroup.GET("", serviceReady(func(c echo.Context) error { return appState.services.thirdPartyHandler.List(c) }))

	itemsGroup := v1.Group("/items")
	itemsGroup.POST("", serviceReady(func(c echo.Context) error { return appState.services.itemHandler.Create(c) }))
	itemsGroup.GET("", serviceReady(func(c echo.Context) error { return appState.services.itemHandler.List(c) }))

	v1.POST("/stock/movements", serviceReady(func(c echo.Context) error { return appState.services.stockHandler.CreateStockMovement(c) }))

	bomGroup := v1.Group("/boms")
	bomGroup.POST("", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.CreateBOM(c) }))
	bomGroup.GET("/:id", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.GetBOMByID(c) }))
	bomGroup.GET("/product/:productID", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.GetBOMByProductID(c) }))
	bomGroup.GET("", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.ListBOMs(c) }))
	bomGroup.PUT("/:id", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.UpdateBOM(c) }))
	bomGroup.DELETE("/:id", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.DeleteBOM(c) }))
	bomGroup.POST("/calculate-cost", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.CalculatePredictiveCost(c) }))
	bomGroup.POST("/produce", serviceReady(func(c echo.Context) error { return appState.services.bomHandler.ProduceItem(c) }))

	marginGroup := v1.Group("/margin")
	marginGroup.GET("/products/:productID", serviceReady(func(c echo.Context) error { return appState.services.marginHandler.GetProductMarginReport(c) }))
	marginGroup.GET("", serviceReady(func(c echo.Context) error { return appState.services.marginHandler.ListOverallMarginReports(c) }))

	invoiceGroup := v1.Group("/invoices")
	invoiceGroup.POST("", serviceReady(func(c echo.Context) error { return appState.services.invoiceHandler.CreateInvoice(c) }))
	invoiceGroup.GET("/:id", serviceReady(func(c echo.Context) error { return appState.services.invoiceHandler.GetInvoice(c) }))

	// Start DB initialization in the background
	go initServices(cfg)

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server failed to start")
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server and PDF Worker Pool...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	pdfWorkerPool.Shutdown(15 * time.Second)

	log.Info().Msg("Server and PDF Worker Pool gracefully stopped.")
}