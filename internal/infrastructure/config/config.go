package config

import (
	"os"
	"strings"

	"service-short-link/internal/domain"

	"github.com/spf13/viper"
)

type configService struct {
	v *viper.Viper
}

// NewConfigService creates a new configuration service
func NewConfigService() domain.ConfigService {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvironmentVariables(v)
	setDefaults(v)

	return &configService{v: v}
}

// bindEnvironmentVariables binds environment variables to config keys
func bindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")
	v.BindEnv("auth.api_keys", "API_KEYS")

	v.BindEnv("server.port", "APP_PORT")

	v.BindEnv("database.host", "DATABASE_HOST")
	v.BindEnv("database.port", "DATABASE_PORT")
	v.BindEnv("database.user", "DATABASE_USER")
	v.BindEnv("database.password", "DATABASE_PASSWORD")
	v.BindEnv("database.dbname", "DATABASE_NAME")

	v.BindEnv("shortlink.test_db_host", "TEST_DB_HOST")
	v.BindEnv("shortlink.test_db_port", "TEST_DB_PORT")
	v.BindEnv("shortlink.test_db_user", "TEST_DB_USER")
	v.BindEnv("shortlink.test_db_password", "TEST_DB_PASSWORD")
	v.BindEnv("shortlink.test_db_name", "TEST_DB_NAME")

	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")

	v.BindEnv("shortlink.base_url", "BASE_URL")
	v.BindEnv("shortlink.shortcode_length", "SHORTCODE_LENGTH")
	v.BindEnv("shortlink.shortcode_min_length", "SHORTCODE_MIN_LENGTH")
	v.BindEnv("shortlink.shortcode_max_length", "SHORTCODE_MAX_LENGTH")
	v.BindEnv("shortlink.default_expiry_seconds", "DEFAULT_EXPIRY_SECONDS")

	v.BindEnv("cache.ttl_links", "CACHE_TTL_LINKS")
	v.BindEnv("cache.ttl_qr", "CACHE_TTL_QR")

	v.BindEnv("e2e.test_mode", "E2E_TEST_MODE")
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "60s")

	v.SetDefault("database.host", "mariadb")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "shortlink")
	v.SetDefault("database.password", "shortlink123")
	v.SetDefault("database.dbname", "shortlink_db")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", "300s")

	// Test database defaults
	v.SetDefault("shortlink.test_db_host", "mariadb_test")
	v.SetDefault("shortlink.test_db_port", 3306)
	v.SetDefault("shortlink.test_db_user", "testuser")
	v.SetDefault("shortlink.test_db_password", "testpass")
	v.SetDefault("shortlink.test_db_name", "shortlink_db_test")

	v.SetDefault("e2e.test_mode", false)

	v.SetDefault("redis.host", "redis")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.pool_size", 50)
	v.SetDefault("redis.min_idle_conns", 10)

	v.SetDefault("auth.jwt_secret", "XPY4b9BSp1FFqu6YhPItyT3qeWSQ9ylB")
	v.SetDefault("auth.api_keys", "dev-api-key")

	v.SetDefault("shortlink.base_url", "http://short.sieuviet.com")
	v.SetDefault("shortlink.shortcode_length", 7)
	v.SetDefault("shortlink.shortcode_min_length", 5)
	v.SetDefault("shortlink.shortcode_max_length", 10)
	// 0 seconds means no expiry by default
	v.SetDefault("shortlink.default_expiry_seconds", 0)

	// Cache defaults
	v.SetDefault("cache.ttl_links", 86400) // 24 hours
	v.SetDefault("cache.ttl_qr", 604800)   // 7 days

	// Rate limiting defaults
	v.SetDefault("rate_limiting.create_links", 60)
	v.SetDefault("rate_limiting.redirect", 300)

	// Performance monitoring defaults
	v.SetDefault("performance.enable_metrics", true)
	v.SetDefault("performance.metrics_interval", "30s")

}

// GetString returns a string configuration value
func (c *configService) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt returns an integer configuration value
func (c *configService) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool returns a boolean configuration value
func (c *configService) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetDuration returns a duration configuration value
func (c *configService) GetDuration(key string) string {
	return c.v.GetString(key)
}

// GetJWTSecret returns the JWT secret key
func (c *configService) GetJWTSecret() string {
	return c.v.GetString("auth.jwt_secret")
}

// GetAPIKeys returns the list of valid API keys
func (c *configService) GetAPIKeys() []string {
	keys := c.v.GetString("auth.api_keys")
	if keys == "" {
		return []string{}
	}
	return strings.Split(keys, ",")
}

// GetDatabaseConfig returns database configuration
func (c *configService) GetDatabaseConfig() domain.DatabaseConfig {
	// Check if E2E test mode is enabled
	if c.v.GetBool("e2e.test_mode") || os.Getenv("E2E_TEST_MODE") == "true" {
		return c.GetTestDatabaseConfig()
	}

	return domain.DatabaseConfig{
		Host:            c.v.GetString("database.host"),
		Port:            c.v.GetInt("database.port"),
		User:            c.v.GetString("database.user"),
		Password:        c.v.GetString("database.password"),
		DBName:          c.v.GetString("database.dbname"),
		MaxOpenConns:    c.v.GetInt("database.max_open_conns"),
		MaxIdleConns:    c.v.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: c.v.GetDuration("database.conn_max_lifetime"),
	}
}

// GetTestDatabaseConfig returns test database configuration
func (c *configService) GetTestDatabaseConfig() domain.DatabaseConfig {
	return domain.DatabaseConfig{
		Host:            c.v.GetString("shortlink.test_db_host"),
		Port:            c.v.GetInt("shortlink.test_db_port"),
		User:            c.v.GetString("shortlink.test_db_user"),
		Password:        c.v.GetString("shortlink.test_db_password"),
		DBName:          c.v.GetString("shortlink.test_db_name"),
		MaxOpenConns:    c.v.GetInt("database.max_open_conns"),
		MaxIdleConns:    c.v.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: c.v.GetDuration("database.conn_max_lifetime"),
	}
}

// GetRedisConfig returns Redis configuration
func (c *configService) GetRedisConfig() domain.RedisConfig {
	return domain.RedisConfig{
		Host:         c.v.GetString("redis.host"),
		Port:         c.v.GetInt("redis.port"),
		Password:     c.v.GetString("redis.password"),
		Database:     c.v.GetInt("redis.database"),
		PoolSize:     c.v.GetInt("redis.pool_size"),
		MinIdleConns: c.v.GetInt("redis.min_idle_conns"),
	}
}

// GetServerConfig returns server configuration
func (c *configService) GetServerConfig() domain.ServerConfig {
	return domain.ServerConfig{
		Port:         c.v.GetInt("server.port"),
		ReadTimeout:  c.v.GetDuration("server.read_timeout"),
		WriteTimeout: c.v.GetDuration("server.write_timeout"),
		IdleTimeout:  c.v.GetDuration("server.idle_timeout"),
	}
}
