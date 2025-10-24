package domain

import "time"

// Configuration types
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// ShortCodeGenerator defines the interface for generating short codes
type ShortCodeGenerator interface {
	GenerateWithLength(length int) string
	IsValid(shortCode string) bool
}

// UniqueCodeService defines the interface for generating unique short codes
type UniqueCodeService interface {
	GenerateUniqueCode(generator ShortCodeGenerator, repo LinkRepository, length int) (string, error)
}

// QRCodeGenerator defines the interface for generating QR codes
type QRCodeGenerator interface {
	Generate(data string, size int) ([]byte, error)
	GenerateWithOptions(data string, size int, options QRCodeOptions) ([]byte, error)
}

// QRCodeOptions represents options for QR code generation
type QRCodeOptions struct {
	Size       int
	BorderSize int
	ErrorLevel QRErrorLevel
	Format     QRFormat
}

// QRErrorLevel represents the error correction level for QR codes
type QRErrorLevel int

const (
	QRErrorLevelLow QRErrorLevel = iota
	QRErrorLevelMedium
	QRErrorLevelQuartile
	QRErrorLevelHigh
)

// QRFormat represents the output format for QR codes
type QRFormat int

const (
	QRFormatPNG QRFormat = iota
	QRFormatJPEG
	QRFormatSVG
)

// URLValidator defines the interface for URL validation
type URLValidator interface {
	IsValid(url string) bool
	Normalize(url string) (string, error)
	IsSafeURL(url string) bool
}

// ConfigService defines the interface for configuration management
type ConfigService interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetDuration(key string) string
	GetAPIKeys() []string
	GetServerConfig() ServerConfig
	GetDatabaseConfig() DatabaseConfig
	GetTestDatabaseConfig() DatabaseConfig
	GetRedisConfig() RedisConfig
}

// LinkUseCase defines the interface for link business logic
type LinkUseCase interface {
	CreateLink(req *CreateLinkRequest, platform, role, channelCode *string) (*CreateLinkResponse, error)
	RedirectLink(shortCode string, trackingData *TrackingData) (string, error)
	GenerateQRCode(shortCode string) ([]byte, error)
}
