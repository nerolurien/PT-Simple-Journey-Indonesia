package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"simplejourney/helpdesk/backend/internal/config"
	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/handler"
	"simplejourney/helpdesk/backend/internal/middleware"
	"simplejourney/helpdesk/backend/internal/notification"
	"simplejourney/helpdesk/backend/internal/repository"
	"simplejourney/helpdesk/backend/internal/seeder"
	"simplejourney/helpdesk/backend/internal/service"
	"simplejourney/helpdesk/backend/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("konfigurasi gagal dimuat: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("peringatan: koneksi database belum tersedia: %v", err)
	} else {
		defer pool.Close()
		if err := migrations.Run(context.Background(), pool); err != nil {
			log.Fatalf("migrasi skema gagal: %v", err)
		}
		log.Println("migrasi skema database berhasil dijalankan")

		if err := seeder.Seed(context.Background(), pool); err != nil {
			log.Printf("peringatan: seeder awal gagal: %v", err)
		} else {
			log.Println("seeder data awal berhasil diverifikasi")
		}
	}

	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	hub := notification.NewHub()

	ticketRepo := repository.NewTicketRepository(pool)
	ticketService := service.NewTicketService(ticketRepo, userRepo, cfg.SLAUrgentHours, cfg.StorageDir, hub)
	ticketHandler := handler.NewTicketHandler(ticketService)

	limiter := middleware.NewIPRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	r := chi.NewRouter()
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(api chi.Router) {
		api.Get("/notifications/stream", hub.HandleStream(cfg.JWTSecret))

		api.Route("/auth", func(auth chi.Router) {
			auth.With(middleware.RateLimit(limiter)).Post("/register", authHandler.Register)
			auth.With(middleware.RateLimit(limiter)).Post("/login", authHandler.Login)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.Authenticate(cfg.JWTSecret))
			protected.Get("/me", authHandler.Me)

			protected.Route("/tickets", func(tickets chi.Router) {
				tickets.Post("/", ticketHandler.Create)
				tickets.Get("/", ticketHandler.List)
				tickets.Get("/{id}", ticketHandler.GetByID)
				tickets.Patch("/{id}", ticketHandler.Update)
				tickets.Post("/{id}/comments", ticketHandler.AddComment)
				tickets.Post("/{id}/attachments", ticketHandler.UploadAttachment)
				tickets.Get("/{id}/attachments/{att_id}", ticketHandler.DownloadAttachment)
				tickets.Delete("/{id}/attachments/{att_id}", ticketHandler.DeleteAttachment)
			})

			protected.Group(func(admin chi.Router) {
				admin.Use(middleware.RequireRoles(domain.RoleAdmin))
				admin.Get("/users", userHandler.List)
				admin.Patch("/users/{id}/role", userHandler.UpdateRole)
				admin.Get("/admin/stats", ticketHandler.GetAdminStats)
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server helpdesk berjalan pada port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server berhenti dengan error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("gagal shutdown server: %v", err)
	}
}
