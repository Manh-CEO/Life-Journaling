# Life Journaling - Go Backend API

Production-ready Go REST API for the Life Journaling application. Users receive weekly email prompts, reply via email, and the system stores their replies as memories. Users can also manage entries manually through a REST API consumed by a Flutter mobile app.

## Tech Stack

- **Language**: Go 1.23
- **HTTP Router**: chi (lightweight, stdlib-compatible)
- **Database**: PostgreSQL (via Supabase) with pgx/v5 driver
- **Migrations**: golang-migrate
- **Config**: envconfig (struct-tag env parsing)
- **Logging**: slog (stdlib, zero deps)
- **Authentication**: JWT (Supabase Auth)
- **Email**: Resend API
- **LLM**: Google Gemini (V2 - V1 stores raw text)
- **Testing**: testify + table-driven tests
- **Containerization**: Docker + docker-compose

## Architecture

Clean Architecture with clear separation of concerns:

```
cmd/api/              # Application entry point
internal/
  domain/             # Business entities & errors
  usecase/            # Business logic & ports (interfaces)
  adapter/            # External integrations
    postgres/         # PostgreSQL repositories
    email/            # Resend email provider
    llm/              # Gemini LLM provider (V1 stub)
  handler/            # HTTP handlers & middleware
    dto/              # Data transfer objects
    middleware/       # Auth, logging, recovery
  config/             # Configuration management
migrations/           # Database migrations
```

## Features

### REST API Endpoints

**Public**
- `GET /health` - Health check

**Authenticated (JWT)**
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update profile/preferences
- `GET /api/v1/memories` - List memories (paginated)
- `POST /api/v1/memories` - Create manual entry
- `GET /api/v1/memories/{id}` - Get single memory
- `PUT /api/v1/memories/{id}` - Update memory
- `DELETE /api/v1/memories/{id}` - Delete memory
- `GET /api/v1/portraits` - List portraits
- `POST /api/v1/portraits` - Create portrait metadata
- `GET /api/v1/portraits/latest` - Latest portrait
- `DELETE /api/v1/portraits/{id}` - Delete portrait

**Internal (API Key)**
- `POST /internal/cron/hourly` - QStash hourly prompt trigger
- `POST /internal/cron/annual` - QStash anchor date trigger
- `POST /internal/webhook/email` - Cloudflare inbound email

### Domain Entities

- **User**: Profile, timezone, prompt schedule, anchor date
- **Memory**: Journal entries with date, location, content, sentiment
- **EngagementLog**: Raw email tracking (pending/completed/failed)
- **Portrait**: Yearly portrait photographs

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & docker-compose
- PostgreSQL (or use docker-compose)
- golang-migrate CLI

### Installation

1. **Clone and setup**
```bash
cd /d/Work/project/UET/Life-Journaling
cp .env.example .env
# Edit .env with your actual credentials
```

2. **Install dependencies**
```bash
go mod download
```

3. **Start database**
```bash
make docker-up
```

4. **Run migrations**
```bash
make migrate-up
```

5. **Run the server**
```bash
make run
# OR build and run binary
make build
./bin/api
```

Server starts on `http://localhost:8080`

### Environment Variables

See `.env.example` for all required variables:
- Database connection (Supabase PostgreSQL)
- JWT secret (Supabase Auth)
- Resend API key (email)
- Gemini API key (LLM)
- QStash signing key (cron)
- Cloudflare webhook secret

## Development

### Make Commands

```bash
make build              # Build the API binary
make run                # Run the API locally
make test               # Run tests with coverage
make test-coverage      # Generate HTML coverage report
make lint               # Run go vet
make clean              # Remove build artifacts
make docker-up          # Start PostgreSQL with docker
make docker-down        # Stop docker services
make migrate-up         # Apply database migrations
make migrate-down       # Rollback migrations
make migrate-create     # Create new migration
make help               # Show all commands
```

### Testing

All services have comprehensive tests with **86.9% coverage** (exceeds 80% requirement):

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test suite
go test ./internal/usecase -v
```

### Database Migrations

Migrations are in `migrations/` directory:

1. **users** - User profiles with prompt schedule
2. **engagement_logs** - Email engagement tracking
3. **memories** - Journal entries
4. **portraits** - Portrait metadata
5. **triggers** - Updated_at triggers

Create new migration:
```bash
make migrate-create name=add_new_table
```

## API Response Format

All endpoints use a consistent envelope:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

Error response:
```json
{
  "success": false,
  "data": null,
  "error": "error message"
}
```

## Authentication

JWT authentication via Supabase:

1. Flutter app authenticates with Supabase Auth
2. Receives JWT token
3. Sends token in `Authorization: Bearer <token>` header
4. Go backend validates JWT and extracts `user_id` + `email`

Internal endpoints use API key authentication (`X-API-Key` header).

## Email Integration

**Outbound (Resend)**
- Weekly prompts sent on user's schedule (day-of-week + hour)
- Annual anchor date emails

**Inbound (Cloudflare Email Routing)**
- Webhook receives email replies
- V1: Saves raw text as pending engagement log
- V2: Will use Gemini API for smart parsing

## Deployment

### Docker Build

```bash
docker build -t life-journaling-api .
docker run -p 8080:8080 --env-file .env life-journaling-api
```

### Production Checklist

- [ ] Set `APP_ENV=production`
- [ ] Use secure `SUPABASE_JWT_SECRET`
- [ ] Enable `DB_SSL_MODE=require`
- [ ] Configure real API keys (Resend, Gemini, QStash)
- [ ] Set up Cloudflare Email Routing webhook
- [ ] Configure QStash cron jobs
- [ ] Set up database backups
- [ ] Enable logging aggregation
- [ ] Configure health check monitoring

## Project Decisions

### V1 Simplifications

1. **LLM Integration**: Stub implementation, saves raw email text. V2 will add Gemini parsing
2. **Portrait Storage**: Metadata only, actual images stored externally (S3, Supabase Storage)
3. **Email Validation**: Basic validation, trust Cloudflare for spam filtering
4. **Authorization**: Simple user ownership checks, can be enhanced in V2

### Design Principles

1. **Immutability**: Services return new objects, never mutate inputs
2. **Clean Architecture**: Clear separation between domain, usecase, adapter, handler
3. **Error Handling**: Domain errors mapped to HTTP status only at handler layer
4. **Pagination**: Standard limit/offset with default 20, max 100
5. **Testing**: Table-driven tests with 80%+ coverage requirement

## API Examples

### Create Memory

```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "entry_date": "2024-02-23",
    "location": "Home",
    "content": "Had a great day working on the Life Journaling project!",
    "sentiment": "positive"
  }'
```

### List Memories

```bash
curl -X GET "http://localhost:8080/api/v1/memories?limit=10&offset=0" \
  -H "Authorization: Bearer <jwt_token>"
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
make docker-up
docker ps

# Verify migrations
make migrate-up
```

### Test Failures

```bash
# Run tests with verbose output
go test ./internal/usecase -v

# Check specific test
go test -run TestUserService_GetByID ./internal/usecase -v
```

### Build Errors

```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

## Contributing

1. Follow Go conventions and project structure
2. Write tests first (TDD) - minimum 80% coverage
3. Use table-driven tests with testify
4. Run `make lint` and `make test` before committing
5. Follow immutability principle - return new objects
6. Document public APIs and complex logic

## License

Proprietary - UET Life Journaling Project

## Support

For questions or issues, contact the development team.
