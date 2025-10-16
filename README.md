# ShortLink Service

A high-performance URL shortening service built with Go, featuring analytics tracking, QR code generation, and dual authentication support.

## Features

ðŸš€ **High Performance**
- Response time < 100ms for redirects
- Support for 1 billion URLs and 100M DAU
- Redis caching with configurable TTL
- Database connection pooling

ðŸ”’ **Security**
- Dual authentication (JWT + API keys)
- URL validation and sanitization
- Rate limiting per endpoint
- Security headers and CORS support

ðŸ“Š **Analytics**
- Real-time click tracking
- Geolocation data (IP-based)
- Device, OS, and browser detection
- Referrer tracking

ðŸ“± **QR Codes**
- On-demand QR code generation
- Optimized PNG output (256x256)
- 7-day caching for performance

ðŸ—ï¸ **Architecture**
- Clean Architecture pattern
- Domain-driven design
- Repository pattern
- Dependency injection

## Tech Stack

- **Language**: Go 1.24
- **Framework**: Gin
- **Database**: MariaDB 11
- **Cache**: Redis 7
- **Proxy**: Nginx
- **Containers**: Docker & Docker Compose

## Quick Start

### Development Setup

1. **Clone and setup**
```bash
cd D:\TungTs\golang\service-short-link
cp .env.example .env
```

2. **Start services**
```bash
# Basic services
docker-compose up -d

# With observability stack
docker-compose --profile observability up -d
```

3. **Access the service**
- API: http://short.sieuviet.com
- Swagger: http://short.sieuviet.com/swagger/index.html
- Grafana: http://localhost:3000 (admin/admin123)
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
  "description": "Description",     // OPTIONAL: Max 1000 characters
  "platform": "web",               // OPTIONAL: Max 50 characters
  "role": "admin",                  // OPTIONAL: Max 50 characters
  "channel_code": "vl24h"          // OPTIONAL: Max 50 characters
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
- `platform`, `role`, `channel_code`: Optional, max 50 characters each


#### Redirect to Original URL
```bash
GET /{shortCode}
Authorization: Bearer {jwt_token}
# OR
X-API-Key: {api_key}
```

#### Generate QR Code
```bash
GET /qr/{shortCode}
Authorization: Bearer {jwt_token}
# OR
X-API-Key: {api_key}
```

### Public APIs (No authentication)

```bash
GET /              # Service info
GET /health        # Health check
GET /swagger/*     # API documentation  
GET /metrics       # Prometheus metrics
```

## Authentication

### JWT Token (Internal Services)
For backend microservices communication:

```json
{
  "platform": "service",    // web | mobile | service
  "role": "employer",       // seeker | employer | mix | system | tracking-email
  "exp": 1735862400,
  "iat": 1733270400
}
```

Usage:
```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### API Key (External Clients)
For external applications:

```bash
X-API-Key: your-api-key-here
# or
apikey: your-api-key-here  # Kong compatible
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `BASE_URL` | Base URL for short links | http://short.sieuviet.com |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | - |
| `API_KEYS` | Comma-separated API keys | - |
| `OBSERVABILITY_ENABLED` | Enable Jaeger tracing | true |
| `METRICS_ENABLED` | Enable Prometheus metrics | true |

### Database Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `MYSQL_HOST` | Database host | localhost |
| `MYSQL_PORT` | Database port | 3306 |
| `MYSQL_USER` | Database user | shortlink |
| `MYSQL_PASSWORD` | Database password | shortlink123 |
| `MYSQL_DATABASE` | Database name | shortlink_db |

### Redis Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `REDIS_PASSWORD` | Redis password | - |

## Architecture

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚   Nginx Proxy   â”‚â”€â”€â”€â–¶â”‚   Go Application â”‚â”€â”€â”€â–¶â”‚    MariaDB      â”‚
â”‚  Rate Limiting  â”‚    â”‚  Clean Architecture â”‚    â”‚   Persistent    â”‚
â”‚  Load Balancing â”‚    â”‚  JWT/API Auth    â”‚    â”‚    Storage      â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜    â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜    â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                â”‚
                                â–¼
                       â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
                       â”‚      Redis      â”‚
                       â”‚   Cache Layer   â”‚
                       â”‚   QR Storage    â”‚
                       â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
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
â”œâ”€â”€ cmd/api/           # Application entry point
â”œâ”€â”€ internal/
â”‚   â”œâ”€â”€ domain/        # Domain entities and interfaces
â”‚   â”œâ”€â”€ usecase/       # Business logic
â”‚   â”œâ”€â”€ infrastructure/# External dependencies
â”‚   â””â”€â”€ handler/       # HTTP handlers and middleware
â”œâ”€â”€ migrations/        # Database migrations
â”œâ”€â”€ config/           # Configuration files
â”œâ”€â”€ nginx/            # Nginx configurations
â”œâ”€â”€ docker/           # Docker configurations
â””â”€â”€ docs/             # API documentation
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
# Build
go build -o shortlink cmd/api/main.go

# Run migrations
migrate -path migrations -database "mysql://..." up

# Start service
./shortlink
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [API Docs](http://short.sieuviet.com/swagger/index.html)
- **Email**: tungts@nhanlucsieuviet.com

---

Built with â¤ï¸ by TungTs
