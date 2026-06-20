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

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/avatar"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/handler"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/storage"
	"github.com/officeryoda/dozingo/internal/worker"
)

const (
	sessionCleanupInterval      = 1 * time.Hour
	tokenCleanupInterval        = 1 * time.Hour
	avatarOrphanCleanupInterval = 1 * time.Hour
	gameAbandonInterval         = 1 * time.Hour
	gameAbandonTimeout          = 6 * time.Hour
	shutdownTimeout             = 10 * time.Second
	defaultAvatarKey            = "default"
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

	totpKey, err := cfg.DecodeTOTPKey()
	if err != nil {
		return fmt.Errorf("decoding TOTP encryption key: %w", err)
	}
	totpCipher, err := auth.NewTOTPCipher(totpKey)
	if err != nil {
		return fmt.Errorf("creating TOTP cipher: %w", err)
	}

	avatarURLs, err := avatar.NewURLBuilder(cfg.GaragePublicURL, cfg.GarageBucketName)
	if err != nil {
		return fmt.Errorf("building avatar URL builder: %w", err)
	}

	garage := storage.NewGarage(ctx, cfg)

	startWorkers(ctx, repos, garage)

	router := createRouter(cfg)
	registerRoutes(ctx, router, repos, pool, cfg, avatarURLs, garage, totpCipher)

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

func startWorkers(ctx context.Context, repos repository.Repos, garage *storage.Garage) {
	worker.NewPeriodic("session_cleanup", sessionCleanupInterval,
		repos.Sessions.DeleteExpiredSessions).Start(ctx)

	worker.NewPeriodic("verification_token_cleanup", tokenCleanupInterval,
		repos.VerificationTokens.DeleteExpired).Start(ctx)

	worker.NewPeriodic("avatar_orphan_cleanup", avatarOrphanCleanupInterval,
		func(ctx context.Context) error {
			return garage.SweepOrphanAvatars(ctx, repos.Users, storage.DefaultSweepConfig())
		}).Start(ctx)

	worker.NewPeriodic("game_abandon", gameAbandonInterval,
		func(ctx context.Context) error {
			n, err := repos.Games.AbandonInactive(ctx, gameAbandonTimeout)
			if err != nil {
				return err
			}
			if n > 0 {
				slog.Info("abandoned inactive games", "count", n)
			}
			return nil
		}).Start(ctx)
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
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := fmt.Fprintf(w, "Dozingo API is running\nDocs: http://localhost:%d/docs", port); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// registerRoutes sets up the Huma API and registers all handler groups.
func registerRoutes(
	ctx context.Context,
	router *chi.Mux,
	repos repository.Repos,
	pool *pgxpool.Pool,
	cfg *config.Config,
	avatarURLs *avatar.URLBuilder,
	garage *storage.Garage,
	totpCipher *auth.TOTPCipher,
) {
	emailSender := email.New(cfg)
	avatarGen := avatar.RandomProfilePicture
	queries := generated.New(pool)
	fallbackURL := avatarURLs.URL(fmt.Sprintf("%s.svg", defaultAvatarKey), "how the fuck did you manage to see this fallback URL")

	ensureDefaultAvatar(ctx, avatarGen, garage)

	humaCfg := huma.DefaultConfig("Dozingo API", "0.2.0")
	humaCfg.DocsPath = "/api/docs"
	api := humachi.New(router, humaCfg)

	apiGroup := huma.NewGroup(api, "/api")
	apiGroup.UseMiddleware(middleware.NewSessionMiddleware(cfg, queries).Handler(api))
	txRunner := repository.NewTxRunner(pool)

	authSvc := service.NewAuth(repos, queries, emailSender, txRunner, avatarGen, garage)
	boardsSvc := service.NewBoards(&repos, queries)
	cellsSvc := service.NewCells(&repos, queries)
	gameCellsSvc := service.NewGameCells(&repos, queries)
	gamesSvc := service.NewGames(&repos, queries, txRunner)
	statsSvc := service.NewStats(&repos, queries)
	twoFASvc := service.NewTwoFactor(&repos, queries, txRunner, totpCipher)
	usersSvc := service.NewUsers(&repos, queries, emailSender, txRunner, garage)
	votesSvc := service.NewVotes(&repos, queries)

	handler.NewAuthHandler(authSvc, avatarURLs, fallbackURL).Register(apiGroup)
	handler.NewBoardsHandler(boardsSvc).Register(apiGroup)
	handler.NewCellsHandler(cellsSvc).Register(apiGroup)
	handler.NewGameCellsHandler(gameCellsSvc).Register(apiGroup)
	handler.NewGamesHandler(gamesSvc).Register(apiGroup)
	handler.NewHealthHandler(pool).Register(api) // Don't use apiGroup here to get around middleware
	handler.NewStatsHandler(statsSvc).Register(apiGroup)
	handler.NewTwoFactor(twoFASvc).Register(apiGroup)
	handler.NewUsersHandler(usersSvc, votesSvc, avatarURLs, fallbackURL).Register(apiGroup)
	handler.NewVotesHandler(votesSvc).Register(apiGroup)

	createOpenAPIFile(api)
}

func ensureDefaultAvatar(ctx context.Context, gen service.AvatarGenerator, uploader storage.ObjectUploader) {
	img, err := gen(defaultAvatarKey)
	if err != nil {
		slog.Warn("failed to generate default avatar", "error", err)
		return
	}
	if err := uploader.Upload(ctx, "default.svg", img); err != nil {
		slog.Warn("failed to upload default avatar", "error", err)
		return
	}
	slog.Info("default avatar uploaded", "key", "default.svg")
}

func createOpenAPIFile(api huma.API) {
	yamlData, _ := api.OpenAPI().YAML()
	err := os.WriteFile("openapi.yaml", yamlData, 0o600)
	if err != nil {
		slog.Warn("failed to write OpenAPI file", "error", err)
	}
}

func createServer(port int, h http.Handler) *http.Server {
	addr := fmt.Sprintf(":%d", port)
	slog.Info("server created", "url", fmt.Sprintf("http://localhost%s", addr))

	return &http.Server{
		Handler:           h,
		Addr:              addr,
		ReadHeaderTimeout: 10 * time.Second,
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
