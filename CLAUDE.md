# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Life Journaling Go backend: REST API for managing personal journal entries via web/mobile app and email. Users receive weekly email prompts, reply via email, and the system stores their responses. Built with Clean Architecture using chi router, PostgreSQL (Supabase), and pgx/v5.

## Architecture

**Clean Architecture with strict layer separation:**

```
Domain Layer (internal/domain/)
  ↓ depends on nothing
UseCase Layer (internal/usecase/)
  ↓ depends on domain via ports (interfaces)
Adapter Layer (internal/adapter/)
  ↓ implements ports
Handler Layer (internal/handler/)
  ↓ HTTP transport, depends on usecase
```

**Dependency flow:** Handler → UseCase → Domain (via interfaces in `usecase/ports.go`)

**Key architectural rules:**
- Domain entities have NO external dependencies (no database, no HTTP)
- UseCase services define ports (interfaces) for external dependencies
- Adapters implement ports (PostgreSQL repos, email providers, LLM providers)
- Handlers only map HTTP ↔ UseCase, domain errors → HTTP status codes at handler layer only

## Dependency Injection Pattern

Dependencies wired in `cmd/api/main.go`:
1. Create repositories (adapters implementing ports)
2. Create external adapters (email, LLM)
3. Inject into services (usecase layer)
4. Inject services into handlers via `RouterDeps` struct

## Core Commands

```bash
# Build and run
make build          # → bin/api
make run            # Run directly with go run
./bin/api           # Run compiled binary

# Testing
make test                           # All tests with coverage report
go test ./internal/usecase -v       # Test specific package
go test -run TestUserService_GetByID ./internal/usecase -v  # Single test
make test-coverage                  # HTML coverage report

# Database
make docker-up                      # Start PostgreSQL container
make migrate-up                     # Apply all migrations
make migrate-down                   # Rollback one migration
make migrate-create name=add_xyz    # Create new migration pair

# Development
make lint           # go vet
make clean          # Remove bin/ and coverage files
go mod tidy         # Clean dependencies
```

## Database Migrations

5 migration pairs in `migrations/` (numbered 000001-000005):
1. users - Profile with prompt schedule (day-of-week + hour in UTC)
2. engagement_logs - Email tracking (pending/completed/failed)
3. memories - Journal entries (with entry_date, location, content, sentiment)
4. portraits - Yearly portrait metadata (storage_path, portrait_year)
5. triggers - Auto-update updated_at columns

Migration naming: `NNNNNN_description.up.sql` and `NNNNNN_description.down.sql`

## Testing Requirements

- **Minimum coverage:** 80% (currently 86.9%)
- **Pattern:** Table-driven tests with testify
- **Mock strategy:** Manual mocks in test files (e.g., `MockUserRepository` implements `IUserRepository`)
- **Test file location:** `*_test.go` in same package as code under test

When adding new usecase service:
1. Define port interfaces in `usecase/ports.go`
2. Implement service with business logic
3. Create `*_test.go` with table-driven tests
4. Mock dependencies using `testify/mock`

## Domain Error Handling

Domain errors defined in `domain/errors.go`:
- `ErrNotFound`, `ErrValidation`, `ErrUnauthorized`, `ErrForbidden`, etc.
- Use `domain.NewDomainError(sentinel, message)` to wrap with context
- Handlers map domain errors to HTTP status codes (see `handler/user_handler.go::handleError`)

**Never return HTTP errors from usecase layer** - return domain errors only.

## Immutability Pattern

**Critical:** All service methods return new objects, never mutate inputs.

Example from `memory_service.go::Update`:
```go
// Build updated memory (immutable: new object)
updated := domain.Memory{
    ID:            existing.ID,
    UserID:        existing.UserID,
    // ... copy all fields
}
// Apply changes to new object
if input.Content != nil {
    updated.Content = *input.Content
}
return s.memoryRepo.Update(ctx, updated)
```

## Authentication Flow

1. Flutter app → Supabase Auth → JWT token
2. Request with `Authorization: Bearer <token>` header
3. `middleware.JWTAuth` validates token, extracts `user_id` + `email`
4. Stores in context: `middleware.UserIDKey`, `middleware.UserEmailKey`
5. Handlers retrieve via `middleware.GetUserID(ctx)`, `middleware.GetUserEmail(ctx)`

Internal endpoints (`/internal/*`) use API key auth via `middleware.APIKeyAuth`.

## API Response Envelope

All endpoints use consistent JSON envelope (defined in `handler/dto/response.go`):
```go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Error   *string     `json:"error"`
    Meta    *Meta       `json:"meta,omitempty"`
}
```

Pagination metadata in `Meta` struct (total, limit, offset).

## Adding New Features

### Add new REST endpoint:
1. Define DTO in `handler/dto/`
2. Add method to existing service OR create new service in `usecase/`
3. Add handler method to appropriate handler (`handler/*_handler.go`)
4. Register route in `handler/router.go`

### Add new repository operation:
1. Add method to port interface in `usecase/ports.go`
2. Implement in `adapter/postgres/*_repo.go`
3. Use in service method

### Add new domain entity:
1. Create in `internal/domain/`
2. Define port interface in `usecase/ports.go`
3. Implement repository in `adapter/postgres/`
4. Create migration in `migrations/`
5. Create service in `usecase/`
6. Create DTOs and handler

## Email Integration

**Outbound (Resend):** `adapter/email/resend.go` implements `IEmailProvider`
- Weekly prompts: `EngagementService.SendHourlyPrompts` (triggered by QStash cron)
- Annual prompts: `EngagementService.SendAnchorDateEmails` (triggered by QStash cron)

**Inbound (Cloudflare Email Routing):** Webhook at `/internal/webhook/email`
- V1: Saves raw email text to `engagement_logs` table (status: pending)
- V2 (future): Will use Gemini API to parse email into structured memory

## LLM Integration (V1 Stub)

`adapter/llm/gemini.go` currently returns raw text as-is. V2 will:
- Parse email content using Gemini API
- Extract structured data (date, location, sentiment)
- Update engagement log status to completed/failed

## Environment Configuration

Required variables (see `.env.example`):
- `APP_PORT`, `APP_ENV`
- `DB_*` - PostgreSQL connection (Supabase)
- `SUPABASE_JWT_SECRET` - For JWT validation
- `RESEND_API_KEY`, `RESEND_FROM_EMAIL` - Email sending
- `GEMINI_API_KEY` - LLM (V2)
- `QSTASH_SIGNING_KEY` - Cron job authentication
- `CF_WEBHOOK_SECRET` - Cloudflare webhook verification

## Pagination

Standard pattern (see `usecase/memory_service.go::normalizePagination`):
- Default limit: 20
- Max limit: 100
- Negative offset: normalized to 0
- Returns `PaginatedResult[T]` with `Items` and `Total`

## Port Definitions

All ports (interfaces) in `usecase/ports.go`:
- `IUserRepository` - User data access
- `IMemoryRepository` - Memory CRUD with pagination
- `IEngagementLogRepository` - Email engagement tracking
- `IPortraitRepository` - Portrait metadata
- `IEmailProvider` - Email sending (Resend)
- `ILLMProvider` - LLM processing (Gemini)

When mocking for tests, implement these interfaces in `*_test.go` files using `testify/mock`.

## Cron Jobs (QStash Integration)

Internal endpoints triggered by QStash:
- `POST /internal/cron/hourly` → `EngagementService.SendHourlyPrompts()`
  - Finds users where `prompt_day_of_week` and `prompt_hour` match current UTC time
  - Sends weekly prompt email to each user

- `POST /internal/cron/annual` → `EngagementService.SendAnchorDateEmails()`
  - Finds users where `anchor_date` (MM-DD) matches today
  - Sends anniversary email with portrait reminder

Both use API key authentication via `middleware.APIKeyAuth`.
