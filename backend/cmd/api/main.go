package main

import (
	"context"
	"errors"
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
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/worker"
)

const (
	sessionCleanupInterval = 1 * time.Hour
	shutdownTimeout        = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	pool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repos := repository.New(pool)

	cleaner := worker.NewSessionCleaner(repos.Sessions, sessionCleanupInterval)
	cleaner.Start(ctx)

	router := createRouter(cfg)
	registerRoutes(router, repos, pool, cfg)

	return serveHTTP(ctx, createServer(cfg.Port, router))
}

// connectDB creates a connection pool to PostgreSQL and verifies the connection.
func connectDB(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	slog.Info("connected to database")
	return pool, nil
}

// createRouter creates a Chi router with standard middleware and a root health page.
func createRouter(cfg *config.Config) *chi.Mux {
	router := chi.NewMux()
	router.Use(chimw.Logger)
	router.Use(chimw.Recoverer)
	router.Get("/", rootHandler(cfg.Port))
	return router
}

func rootHandler(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := fmt.Fprintf(w, "Dozingo API is running\nDocs: http://localhost:%d/docs", port); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// registerRoutes sets up the Huma API and registers all handler groups.
func registerRoutes(router *chi.Mux, repos repository.Repos, pool *pgxpool.Pool, cfg *config.Config) {
	queries := generated.New(pool)

	api := humachi.New(router, huma.DefaultConfig("Dozingo API", "0.2.0"))
	api.UseMiddleware(middleware.NewSessionMiddleware(cfg, queries).Handler(api))

	apiGroup := huma.NewGroup(api, "/api")
	txRunner := repository.NewTxRunner(pool)

	boardsSvc := service.NewBoards(repos.Boards, queries)
	cellsSvc := service.NewCells(repos.Cells, repos.Boards, queries)
	gameCellsSvc := service.NewGameCells(repos.GameCells, repos.Games, queries)
	gamesSvc := service.NewGames(repos.Games, queries)
	votesSvc := service.NewVotes(repos.Votes, queries)
	authSvc := service.NewAuth(repos, queries, txRunner)

	handler.NewHealthHandler().Register(apiGroup)
	handler.NewBoardsHandler(boardsSvc).Register(apiGroup)
	handler.NewCellsHandler(cellsSvc).Register(apiGroup)
	handler.NewGameCellsHandler(gameCellsSvc).Register(apiGroup)
	handler.NewGamesHandler(gamesSvc).Register(apiGroup)
	handler.NewVotesHandler(votesSvc).Register(apiGroup)
	handler.NewAuthHandler(authSvc).Register(apiGroup)
}

func createServer(port int, handler http.Handler) *http.Server {
	addr := fmt.Sprintf(":%d", port)
	slog.Info("server created", "url", fmt.Sprintf("http://localhost%s", addr))
	return &http.Server{
		Handler: handler,
		Addr:    addr,
	}
}

// serveHTTP starts the server and blocks until the context is cancelled or the
// server fails. On context cancellation it attempts a graceful shutdown.
func serveHTTP(ctx context.Context, srv *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- startServer(srv)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("server shut down cleanly")
	return nil
}

// startServer runs the server and returns nil on clean shutdown,
// or the underlying error for unexpected failures.
func startServer(srv *http.Server) error {
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
