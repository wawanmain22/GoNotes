package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gonotes/internal/config"
	"gonotes/internal/handler"
	"gonotes/internal/middleware"
	"gonotes/internal/repository"
	"gonotes/internal/service"
	"gonotes/internal/utils"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	db, err := utils.ConnectDB(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connected successfully")

	// Initialize Redis
	redisClient, err := utils.ConnectRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	// Initialize validator
	validator := utils.NewValidator()

	// Initialize services
	userService := service.NewUserServiceWithRedis(userRepo, redisClient)
	sessionService := service.NewSessionService(sessionRepo, userRepo, redisClient, cfg)
	noteService := service.NewNoteService(noteRepo, userRepo, validator)

	// Initialize audit service
	auditService := service.NewAuditService()
	defer auditService.Close()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, sessionService)
	noteHandler := handler.NewNoteHandler(noteService)
	sessionHandler := handler.NewSessionHandler(sessionService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionService, cfg)
	rateLimitConfig := middleware.DefaultRateLimitConfig(redisClient)

	// Setup routes
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.CORSMiddleware)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.AuditLogMiddleware())
	r.Use(middleware.RateLimitMiddleware(rateLimitConfig))
	r.Use(middleware.DDoSProtectionMiddleware(redisClient))
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Public routes (authentication)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
		r.Post("/logout", authHandler.Logout)
	})

	// Protected routes (require authentication)
	r.Route("/api/v1/user", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		// Profile management
		r.Get("/profile", authHandler.GetProfile)
		r.Put("/profile", authHandler.UpdateProfile)

		// Basic session info (legacy)
		r.Get("/sessions", authHandler.GetSessions)
	})

	// Advanced session management routes
	r.Route("/api/v1/user/sessions", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		// Get all active sessions with device info
		r.Get("/active", sessionHandler.GetActiveSessions)

		// Session statistics
		r.Get("/stats", sessionHandler.GetSessionsStats)

		// Invalidate all sessions (logout from all devices)
		r.Delete("/", sessionHandler.InvalidateAllSessions)

		// Invalidate specific session (logout from specific device)
		r.Delete("/{sessionId}", sessionHandler.InvalidateSession)

		// Alternative endpoint for session invalidation via POST
		r.Post("/invalidate", sessionHandler.InvalidateSessionByRequest)
	})

	// Public notes routes (no authentication required)
	r.Route("/api/v1/notes", func(r chi.Router) {
		// Public endpoints
		r.Get("/public", noteHandler.GetPublicNotes)

		// Protected endpoints (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			// Basic CRUD operations
			r.Post("/", noteHandler.CreateNote)
			r.Get("/", noteHandler.GetNotes)
			r.Get("/{id}", noteHandler.GetNote)
			r.Put("/{id}", noteHandler.UpdateNote)
			r.Delete("/{id}", noteHandler.DeleteNote)

			// Advanced operations
			r.Post("/search", noteHandler.SearchNotes)
			r.Post("/bulk", noteHandler.BulkUpdateNotes)
			r.Get("/stats", noteHandler.GetNoteStats)
			r.Get("/tags", noteHandler.GetUserTags)
			r.Get("/tag/{tag}", noteHandler.GetNotesByTag)

			// Note-specific operations
			r.Post("/{id}/restore", noteHandler.RestoreNote)
			r.Delete("/{id}/hard", noteHandler.HardDeleteNote)
			r.Post("/{id}/duplicate", noteHandler.DuplicateNote)
			r.Post("/{id}/toggle-public", noteHandler.ToggleNotePublicStatus)
		})
	})

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Starting server on port %s", cfg.AppPort)
	log.Printf("Health check: http://localhost%s/health", serverAddr)
	log.Printf("API Documentation: http://localhost%s/api/v1", serverAddr)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Cleanup
	log.Println("Server stopped")
}
