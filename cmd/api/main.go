package main

import (
    "context"
    "database/sql"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

	"service-short-link/internal/domain"
	"service-short-link/internal/handler"
	"service-short-link/internal/handler/middleware"
	"service-short-link/internal/infrastructure/cache"
	"service-short-link/internal/infrastructure/config"
	"service-short-link/internal/infrastructure/database"
	"service-short-link/internal/infrastructure/services"
	"service-short-link/internal/usecase"
	"service-short-link/pkg/logger"

    _ "service-short-link/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	serviceName    = "shortlink-service"
	serviceVersion = "1.0.0"
)

// @title ShortLink Service API
// @version 1.0.0
// @description High-performance URL shortening service with analytics tracking

// @contact.name TungTs
// @contact.email tungts@nhanlucsieuviet.com

// @host short.vieclam24h.vn
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token for internal services

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key for external clients

func main() {
	if err := logger.InitLogger(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Short Link Service", map[string]interface{}{
		"version": "1.0.0",
		"env":     os.Getenv("APP_ENV"),
		"pid":     os.Getpid(),
		"working_dir": func() string {
			wd, _ := os.Getwd()
			return wd
		}(),
	})
	
	configService := config.NewConfigService()
	db, err := initDatabase(configService)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "Failed to initialize database", "error_type=database_init_failed")
		os.Exit(1)
	}
	defer db.Close()

	// Run database migrations
	if err := runMigrations(db, configService); err != nil {
		logger.Warn("Failed to run migrations", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize Redis
	redisClient, err := initRedis(configService)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "Failed to initialize Redis", "error_type=redis_init_failed")
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize services
	var linkRepo domain.LinkRepository
	var analyticsRepo domain.AnalyticsRepository
	var linkCache domain.LinkCache

	// Create base repositories
	baseLinkRepo := database.NewLinkRepository(db)
	baseAnalyticsRepo := database.NewAnalyticsRepository(db)
	baseLinkCache := cache.NewLinkCache(redisClient)

    linkRepo = baseLinkRepo
    analyticsRepo = baseAnalyticsRepo
    linkCache = baseLinkCache
	
    minLen := configService.GetInt("shortlink.shortcode_min_length")
    maxLen := configService.GetInt("shortlink.shortcode_max_length")
    shortCodeGenerator := services.NewShortCodeGeneratorWithBounds(minLen, maxLen)
	qrGenerator := services.NewQRCodeGenerator()
	urlValidator := services.NewURLValidator()
	userAgentParser := services.NewUserAgentParser()

	linkUseCase := usecase.NewLinkUseCase(
		linkRepo,
		linkCache,
		analyticsRepo,
		shortCodeGenerator,
		qrGenerator,
		urlValidator,
		userAgentParser,
		configService,
	)

	linkHandler := handler.NewLinkHandler(linkUseCase)
	healthHandler := handler.NewHealthHandler(serviceVersion)

	authMiddleware := middleware.NewAuthMiddleware(
		configService.GetString("auth.jwt_secret"),
		configService.GetAPIKeys(),
	)

    router := setupRouter(configService, authMiddleware, linkHandler, healthHandler)

	// Create server
	serverConfig := configService.GetServerConfig()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverConfig.Port),
		Handler:      router,
		ReadTimeout:  serverConfig.ReadTimeout,
		WriteTimeout: serverConfig.WriteTimeout,
		IdleTimeout:  serverConfig.IdleTimeout,
	}

	go func() {
		logger.Info("Starting server", map[string]interface{}{
			"port": serverConfig.Port,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorWithCockroachSimple(err, "Server failed to start", "error_type=server_start_failed")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", nil)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.ErrorWithCockroachSimple(err, "Server forced to shutdown", "error_type=server_shutdown_failed")
	}

	logger.Info("Server shutdown complete", nil)
}

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(
    configService domain.ConfigService,
    authMiddleware *middleware.AuthMiddleware,
    linkHandler *handler.LinkHandler,
    healthHandler *handler.HealthHandler,
) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.AppErrorLogger())
	router.Use(middleware.CORS())
	router.Use(middleware.SecurityHeaders())

	// Public routes (no authentication required)
	router.GET("/", healthHandler.GetServiceInfo)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessProbe)
	router.GET("/live", healthHandler.LivenessProbe)

	// Swagger documentation
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))



	// Protected routes (require JWT or API Key) - Only for management APIs
	protected := router.Group("/")
	protected.Use(authMiddleware.FlexibleAuth())
	{
		api := protected.Group("/api/v1")
		{
			api.POST("/links", linkHandler.CreateLink)
		}
	}

	// Public routes (no authentication required) - For short link access
	router.GET("/:code", linkHandler.RedirectLink)
	router.GET("/qr/:code", linkHandler.GenerateQRCode)

	return router
}

// initDatabase initializes the database connection
func initDatabase(configService domain.ConfigService) (*sql.DB, error) {
	dbConfig := configService.GetDatabaseConfig()
	
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=false&allowNativePasswords=true&multiStatements=false&interpolateParams=true&readTimeout=5s&writeTimeout=5s&timeout=5s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

    if err := pingDatabaseWithRetry(db, 8, 500*time.Millisecond, 5*time.Second); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

	return db, nil
}

// pingDatabaseWithRetry pings the DB with exponential backoff up to maxAttempts
func pingDatabaseWithRetry(db *sql.DB, maxAttempts int, initialBackoff time.Duration, maxBackoff time.Duration) error {
    backoff := initialBackoff
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        err := db.PingContext(ctx)
        cancel()
        if err == nil {
            return nil
        }
        if attempt == maxAttempts {
            return err
        }
        time.Sleep(backoff)
        backoff *= 2
        if backoff > maxBackoff {
            backoff = maxBackoff
        }
    }
    return nil
}

// initRedis initializes the Redis connection
func initRedis(configService domain.ConfigService) (*redis.Client, error) {
	redisConfig := configService.GetRedisConfig()
	
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.Database,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB, configService domain.ConfigService) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully", nil)
	return nil
}

