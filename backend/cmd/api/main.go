package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/handler"
	"github.com/officeryoda/dozingo/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("failed to load config", "error", err)
	}

	pool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	defer pool.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Periodically remove expired sessions
	handler.StartSessionCleanup(ctx, generated.New(pool))

	router := createRouter(cfg)
	registerRoutes(router, pool)
	srv := createServer(cfg.Port, router)
	go startServer(srv)

	<-ctx.Done() // block until SIGTERM / Ctrl-C

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx) // drain in-flight requests
}

// connectDB creates a connection pool to PostgreSQL and verifies the connection.
func connectDB(databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	slog.Info("Connected to database")
	return pool, nil
}

// createRouter creates a Chi router with standard middleware and a root health page.
func createRouter(cfg *config.Config) *chi.Mux {
	router := chi.NewMux()
	router.Use(chimw.Logger)
	router.Use(chimw.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := fmt.Fprintf(w, "Dozingo API is running\nDocs: http://localhost:%d/docs", cfg.Port); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	return router
}

// registerRoutes sets up the Huma API and registers all handler groups.
func registerRoutes(router *chi.Mux, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	api := humachi.New(router, huma.DefaultConfig("Dozingo API", "0.2.0"))
	api.UseMiddleware(middleware.SessionUser(api, queries))

	apiGroup := huma.NewGroup(api, "/api")

	handler.RegisterHealth(apiGroup)
	handler.RegisterBoards(apiGroup, pool)
	handler.RegisterCells(apiGroup, pool)
	handler.RegisterVotes(apiGroup, pool)
	handler.RegisterGames(apiGroup, pool)
	handler.RegisterGameCells(apiGroup, pool)
	handler.RegisterAuth(apiGroup, pool)
}

func createServer(port int, handler http.Handler) *http.Server {
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Handler: handler, Addr: addr}

	slog.Info("Server created", "addr", addr)
	return srv
}

func startServer(srv *http.Server) {
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "error", err)
	}
}
