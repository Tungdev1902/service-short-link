# ShortLink Service

A high-performance URL shortening service built with Go, featuring analytics tracking, QR code generation, and dual authentication support.

## Features

🚀 **High Performance**
- Response time < 100ms for redirects with Redis caching
- Support for 1+ billion URLs with optimized database indexes
- Redis caching with configurable TTL (24h links, 7d QR codes)
- Database connection pooling with timeout handling
- Async analytics processing for better performance

🔐 **Security**
- Dual authentication (JWT tokens + API keys)
- URL validation and safety checks (blocks private IPs, malicious patterns)
- Nginx rate limiting and proxy protection
- Security headers and CORS support
- JWT claims validation with expiry checks

📊 **Analytics**
- Real-time click tracking with async processing
- User agent parsing (device, OS, browser detection)
- Referrer tracking and IP address logging
- Click count and last accessed timestamp
- Structured logging for monitoring

📱 **QR Codes**
- On-demand QR code generation (256x256 PNG)
- Redis caching for 7-day performance
- Optimized with configurable size and error correction

🗃️ **Architecture**
- Clean Architecture with domain-driven design
- Repository pattern with interface segregation
- Dependency injection and service abstraction
- Error handling with custom domain errors
- Configuration service for centralized settings

## Tech Stack

- **Language**: Go (with Gin web framework)
- **Database**: MariaDB with optimized indexes
- **Cache**: Redis 7 with connection pooling
- **Proxy**: Nginx with rate limiting
- **Containers**: Docker & Docker Compose
- **Monitoring**: Prometheus, Grafana, Jaeger (optional)
- **Documentation**: Swagger/OpenAPI

## Quick Start

### Development Setup

1. **Clone and setup**
```bash
git clone https://github.com/Tungdev1902/service-short-link.git
cd service-short-link
cp .env.example .env
# Edit .env with your configuration values
```

2. **Start services**
```bash
# Basic services
docker-compose up -d

# With observability stack
docker-compose --profile observability up -d
```

3. **Access the service**
- API: http://short.vieclam24h.vn
- Swagger: http://short.vieclam24h.vn/swagger/index.html
- Grafana: http://localhost:3000 (admin/admin password from env)
- Prometheus: http://localhost:9090
- Jaeger: http://localhost:16686

### Production Setup

1. **Configure environment**
```bash
cp .env.production .env
# Edit .env with your production values
```

2. **Deploy**
```bash
docker-compose -f docker-compose.production.yml up -d
```

## API Endpoints

### Protected APIs (JWT or API Key required)

#### Create Short Link
```bash
POST /api/v1/links
Content-Type: application/json
Authorization: Bearer {jwt_token}
# OR
X-API-Key: {api_key}

{
  "original_url": "https://example.com/very/long/path",  // REQUIRED: Valid URL (max 2048 chars)
  "short_code": "abc123",           // OPTIONAL: 5-10 alphanumeric chars, not reserved words
  "title": "Example Website",       // OPTIONAL: Max 255 characters
  "description": "Description"      // OPTIONAL: Max 1000 characters
}
```

**Validation Rules:**
- `original_url`: Required, must be valid URL format, max 2048 characters
- `short_code`: Optional, if provided:
  - Must be 5-10 characters long
  - Only alphanumeric characters (a-z, A-Z, 0-9)
  - Cannot be reserved words (admin, api, swagger, health, etc.)
  - Must be unique (not already exists)
- `title`: Optional, max 255 characters
- `description`: Optional, max 1000 characters
- `expires_at`: **Automatically set** to default expiry from config (365 days from creation)
- `platform`, `role`, `channel_code`: **Automatically extracted from JWT token** for internal services

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "short_code": "aB3x9K2",
    "short_url": "http://short.vieclam24h.vn/aB3x9K2",
    "qr_url": "http://short.vieclam24h.vn/qr/aB3x9K2",
    "original_url": "https://example.com/very/long/path",
    "title": "Example Website",
    "description": "Description",
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2024-01-15T10:30:45Z"
  }
}
```

### Public APIs (No authentication required)

#### Redirect to Original URL
```bash
GET /{shortCode}    # Redirects with 302 status + analytics tracking
```

#### Generate QR Code
```bash
GET /qr/{shortCode} # Returns PNG image (256x256)
```

#### Service Information
```bash
GET /              # Service info and version
GET /health        # Health check
GET /ready         # Readiness probe
GET /live          # Liveness probe
GET /swagger/*     # API documentation
```

## Authentication

### JWT Token (Internal Services)
For backend microservices communication. JWT payload should include:

```json
{
  "platform": "web",         // web | mobile | service
  "role": "employer",        // seeker | employer | mix | system | tracking-email
  "channel_code": "vl24h",   // Optional: channel identifier
  "exp": 1735862400,         // Expiration timestamp
  "iat": 1733270400          // Issued at timestamp
}
```

Usage:
```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Note**: `platform`, `role`, and `channel_code` from JWT claims are automatically extracted and stored with the link for tracking purposes.

### API Key (External Clients)
For external applications and third-party integrations:

```bash
X-API-Key: your-api-key-here
# or
apikey: your-api-key-here  # Kong compatible
```

External API key usage defaults to:
- `platform`: "external"
- `role`: "client"
- `is_internal`: false

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `APP_PORT` | Application port | 8080 |
| `DOMAIN` | Service domain | short.vieclam24h.vn |
| `BASE_URL` | Base URL for short links | http://short.vieclam24h.vn |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | - |
| `API_KEYS` | Comma-separated API keys | - |
| `OBSERVABILITY_ENABLED` | Enable Jaeger tracing | true |
| `METRICS_ENABLED` | Enable Prometheus metrics | true |

### Database Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_HOST` | Database host | mariadb |
| `DATABASE_PORT` | Database port | 3306 |
| `DATABASE_NAME` | Database name | - |
| `DATABASE_USER` | Database user | - |
| `DATABASE_PASSWORD` | Database password | - |
| `MYSQL_ROOT_PASSWORD` | Root password for container | - |
| `MYSQL_DATABASE` | Database name for container | - |
| `MYSQL_USER` | Database user for container | - |
| `MYSQL_PASSWORD` | Database password for container | - |

### Redis Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis host | redis |
| `REDIS_PORT` | Redis port | 6379 |
| `REDIS_PASSWORD` | Redis password | - |

### Performance Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `CACHE_TTL_LINKS` | Links cache TTL (seconds) | 86400 (24h) |
| `CACHE_TTL_QR` | QR codes cache TTL (seconds) | 604800 (7d) |
| `SHORTCODE_LENGTH` | Default short code length | 7 |
| `SHORTCODE_MIN_LENGTH` | Min short code length | 5 |
| `SHORTCODE_MAX_LENGTH` | Max short code length | 10 |
| `DEFAULT_EXPIRY_DAYS` | Default link expiry days | 365 |

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Nginx Proxy   │───▶│   Go Application │───▶│    MariaDB      │
│  Rate Limiting  │    │  Clean Architecture │    │   Persistent    │
│  Load Balancing │    │  JWT/API Auth    │    │    Storage      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │      Redis      │
                       │   Cache Layer   │
                       │   QR Storage    │
                       └─────────────────┘
```

### Layers

1. **Domain Layer**: Entities, interfaces, business rules
2. **Use Case Layer**: Business logic, orchestration
3. **Infrastructure Layer**: Database, cache, external services
4. **Handler Layer**: HTTP handlers, middleware

## Performance

### Benchmarks
- **Redirect**: ~10-15ms (cached), ~50-80ms (DB)
- **Create Link**: ~100-150ms
- **QR Generation**: ~5ms (cached), ~200ms (generate)

### Scalability
- **URLs**: Supports 1+ billion URLs
- **DAU**: 100+ million daily active users
- **RPS**: 10,000+ requests per second
- **Cache Hit Rate**: >95% for redirects

### Optimizations
- Redis caching (24h links, 7d QR codes)
- Database connection pooling
- Async analytics processing
- Nginx rate limiting
- Optimized database indexes

## Security

### URL Validation
- Protocol validation (HTTP/HTTPS only)
- Domain validation
- Private IP blocking
- Malicious pattern detection

### Rate Limiting
- API endpoints: 30-60 req/min
- Redirects: 300-1000 req/min
- Configurable per environment

### Headers
- CORS support
- Security headers (XSS, CSP, etc.)
- HSTS in production

## Monitoring

### Metrics (Prometheus)
- Request rate and latency
- Error rates
- Cache hit rates
- Database connection pool

### Logging
- Structured JSON logs
- Access logs via Nginx
- Error tracking
- Performance monitoring

### Tracing (Jaeger)
- Distributed tracing
- Request flow visualization
- Performance bottleneck identification

## Development

### Code Structure
```
├── cmd/api/                    # Application entry point
│   └── main.go                # Main application with dependency injection
├── internal/
│   ├── domain/                # Domain entities and interfaces
│   │   ├── analytics.go       # Analytics domain model
│   │   ├── errors.go          # Custom domain errors
│   │   ├── link.go            # Link domain model and interfaces
│   │   └── services.go        # Service interfaces and config types
│   ├── usecase/               # Business logic layer
│   │   └── link_usecase.go    # Link business logic and orchestration
│   ├── infrastructure/        # External dependencies
│   │   ├── cache/             # Redis cache implementation
│   │   ├── config/            # Configuration service
│   │   ├── database/          # Database repositories
│   │   └── services/          # External service implementations
│   └── handler/               # HTTP layer
│       ├── middleware/        # Authentication, CORS, logging middleware
│       ├── transform/         # DTO transformations
│       ├── link_handler.go    # Link HTTP handlers
│       ├── health_handler.go  # Health check handlers
│       └── response.go        # Response helpers
├── migrations/                # Database migrations
│   ├── 000001_create_links_table.up.sql
│   ├── 000001_create_links_table.down.sql
│   ├── 000002_create_analytics_table.up.sql
│   └── 000002_create_analytics_table.down.sql
├── nginx/                     # Nginx configurations
│   ├── local.conf            # Development config
│   ├── production.conf       # Production config
│   └── ssl/                  # SSL certificates
├── docker/                    # Docker configurations
│   └── redis/                # Redis configuration
├── docs/                      # API documentation
│   ├── docs.go               # Swagger generated docs
│   ├── swagger.json          # OpenAPI spec
│   └── swagger.yaml          # OpenAPI spec
├── pkg/                       # Shared packages
│   └── logger/               # Structured logging
├── scripts/                   # Utility scripts
│   ├── build-production.sh   # Production build script
│   ├── dev-setup.sh          # Development setup
│   └── generate-swagger.sh   # Swagger generation
└── vendor/                    # Go module dependencies
```

### Running Tests
```bash
go test ./...
go test -race ./...
go test -cover ./...
```

### Code Generation
```bash
# Generate swagger docs
swag init -g cmd/api/main.go

# Generate mocks
mockgen -source=internal/domain/interfaces.go
```

## Deployment

### Docker
```bash
# Development
docker-compose up -d

# Production
docker-compose -f docker-compose.production.yml up -d
```

### Manual Deployment
```bash
# Build binary
go build -o shortlink cmd/api/main.go

# Run database migrations
migrate -path migrations -database "mysql://user:password@tcp(host:port)/database" up

# Set environment variables and start service
export JWT_SECRET="your-jwt-secret-here"
export API_KEYS="your-api-keys-here"
./shortlink
```

## API Flow Details

### Create Short Link Flow

1. **Request Validation**
   - JSON binding and input validation
   - URL format and length checks (max 2048 chars)
   - Short code validation (5-10 alphanumeric chars if provided)
   - Title/description length limits

2. **Authentication**
   - JWT token validation and claims extraction
   - API key validation for external clients
   - Platform/role/channel_code extraction from JWT

3. **Business Logic Processing**
   - URL normalization and safety validation
   - Short code generation (if not provided) with uniqueness check
   - Race condition handling for code generation
   - Default expiry calculation (365 days from config)

4. **Data Persistence**
   - Database insertion with connection timeout (5s)
   - Auto-increment ID assignment
   - Error handling for duplicate short codes

5. **Caching**
   - Redis cache storage with configurable TTL (24h default)
   - JSON serialization of link data
   - Non-blocking cache operations (failures logged but don't fail request)

6. **Response Generation**
   - Build complete response with URLs
   - QR code URL generation
   - DTO transformation and JSON response

### Redirect Flow

1. **Cache Lookup** - Fast Redis lookup for link data
2. **Database Fallback** - Query database if cache miss
3. **Validation** - Check if link is active and not expired
4. **Analytics Tracking** - Async user agent parsing and analytics recording
5. **Redirect** - HTTP 302 redirect to original URL

## Error Handling

The API uses standard HTTP status codes and returns structured error responses:

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request body",
    "details": "original_url is required"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid input data or malformed request |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Link is inactive or access denied |
| `NOT_FOUND` | 404 | Short link not found |
| `CONFLICT` | 409 | Short code already exists |
| `GONE` | 410 | Link has expired |
| `INTERNAL_ERROR` | 500 | Server error or database issues |

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [API Docs](http://short.vieclam24h.vn/swagger/index.html)
- **Repository**: [GitHub](https://github.com/Tungdev1902/service-short-link)
- **Email**: tungts@nhanlucsieuviet.com

---

Built with ❤️ by TungTs