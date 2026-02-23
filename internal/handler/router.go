package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/life-journaling/core/internal/config"
	"github.com/life-journaling/core/internal/handler/dto"
	"github.com/life-journaling/core/internal/handler/middleware"
	"github.com/life-journaling/core/internal/usecase"
)

// RouterDeps holds all dependencies needed by the router.
type RouterDeps struct {
	Config            config.Config
	UserService       *usecase.UserService
	MemoryService     *usecase.MemoryService
	PortraitService   *usecase.PortraitService
	EngagementService *usecase.EngagementService
	IngestionService  *usecase.IngestionService
}

// NewRouter creates and configures the HTTP router.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, dto.NewSuccessResponse(map[string]string{
			"status": "ok",
		}))
	})

	// Handlers
	userHandler := NewUserHandler(deps.UserService)
	memoryHandler := NewMemoryHandler(deps.MemoryService)
	portraitHandler := NewPortraitHandler(deps.PortraitService)
	cronHandler := NewCronHandler(deps.EngagementService)
	webhookHandler := NewWebhookHandler(deps.IngestionService)

	// Authenticated API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.JWTAuth(deps.Config.Supabase.JWTSecret))

		// Users
		r.Get("/users/me", userHandler.GetMe)
		r.Put("/users/me", userHandler.UpdateMe)

		// Memories
		r.Get("/memories", memoryHandler.List)
		r.Post("/memories", memoryHandler.Create)
		r.Get("/memories/{id}", memoryHandler.GetByID)
		r.Put("/memories/{id}", memoryHandler.Update)
		r.Delete("/memories/{id}", memoryHandler.Delete)

		// Portraits
		r.Get("/portraits", portraitHandler.List)
		r.Post("/portraits", portraitHandler.Create)
		r.Get("/portraits/latest", portraitHandler.GetLatest)
		r.Delete("/portraits/{id}", portraitHandler.Delete)
	})

	// Internal routes (API key auth)
	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(deps.Config.QStash.SigningKey))

		r.Post("/cron/hourly", cronHandler.Hourly)
		r.Post("/cron/annual", cronHandler.Annual)
	})

	// Webhook routes (signature verification in handler)
	r.Post("/internal/webhook/email", webhookHandler.InboundEmail)

	return r
}
