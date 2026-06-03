package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/generated"
)

type contextKey int

const (
	contextSessionUser contextKey = iota
	contextSessionSlot
	contextHumaCtx
	contextSessionMW
)

const (
	cookieSessionToken        = "session_token"
	sessionTokenTTL           = 30 * 24 * time.Hour
	sessionExtensionThreshold = 7 * 24 * time.Hour
)

// sessionSlot is a mutable holder shared across a single request so that a
// handler calling RequireSession multiple times reuses the same minted
// session instead of inserting again.
type sessionSlot struct {
	row    generated.GetSessionUserByTokenRow
	filled bool
}

// SessionMiddleware bundles the dependencies needed by the session-aware
// middleware and its companion cookie helpers (RequireSession,
// ClearSessionTokenCookie). Construct it with NewSessionMiddleware and pass
// the result into huma's UseMiddleware.
type SessionMiddleware struct {
	cfg     *config.Config
	queries *generated.Queries
}

// NewSessionMiddleware returns a SessionMiddleware bound to the given config
// and query handle. The config drives cookie behaviour (Secure flag); the
// queries handle is used to look up sessions during request handling.
func NewSessionMiddleware(cfg *config.Config, queries *generated.Queries) *SessionMiddleware {
	if cfg == nil {
		panic("middleware: nil *config.Config")
	}
	if queries == nil {
		panic("middleware: nil *generated.Queries")
	}
	return &SessionMiddleware{cfg: cfg, queries: queries}
}

// Handler is a read-only middleware. If the request carries a valid
// session_token cookie, it loads the session, optionally extends it, and
// stashes it in the request context. Requests without a cookie do not touch
// the database. Handlers that need to persist anonymous activity should call
// RequireSession to lazily mint a session.
func (m *SessionMiddleware) Handler(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Always attach an empty slot so RequireSession can cache a minted
		// session within the lifetime of this request.
		slot := &sessionSlot{}
		ctx = huma.WithValue(ctx, contextSessionSlot, slot)

		// Stash the huma.Context itself so typed handlers (which only see
		// context.Context) can reach AppendHeader/ReadCookie via the
		// RequireSessionCtx helper.
		ctx = huma.WithValue(ctx, contextHumaCtx, ctx)

		// Stash the middleware itself so the free-function helpers
		// (RequireSession, ClearSessionTokenCookie) can read config and
		// queries from the request context without package-level globals.
		ctx = huma.WithValue(ctx, contextSessionMW, m)

		sessionToken := getSessionTokenFromCookie(ctx)
		if sessionToken == "" {
			next(ctx)
			return
		}

		// The cookie carries plaintext; the DB stores SHA-256 hex.
		sessionUser, err := m.queries.GetSessionUserByToken(ctx.Context(), auth.HashToken(sessionToken))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ClearSessionTokenCookie(ctx.Context())
				next(ctx)
				return
			}
			slog.Error("failed to look up session", "error", err)
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error")
			return
		}

		extendCloseToExpiredSessions(&sessionUser, m.queries, ctx)

		slot.row = sessionUser
		slot.filled = true
		ctx = huma.WithValue(ctx, contextSessionUser, sessionUser)
		next(ctx)
	}
}

// SessionUserFromContext returns the session previously set by SessionUser
// middleware. The second return value is false when the request had no valid
// session cookie. Handlers that may need to mint a session for anonymous
// users should call RequireSession instead.
func SessionUserFromContext(ctx context.Context) (generated.GetSessionUserByTokenRow, bool) {
	if slot, ok := ctx.Value(contextSessionSlot).(*sessionSlot); ok && slot.filled {
		return slot.row, true
	}
	v, ok := ctx.Value(contextSessionUser).(generated.GetSessionUserByTokenRow)
	return v, ok
}

// RequireSession returns the request's session, minting and setting a
// Set-Cookie header for an anonymous one if none exists yet.
func RequireSession(ctx context.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	humaCtx, ok := humaContextFrom(ctx)
	if !ok {
		return generated.GetSessionUserByTokenRow{}, errors.New("RequireSession called without Session middleware")
	}

	mw, ok := sessionMiddlewareFrom(ctx)
	if !ok {
		return generated.GetSessionUserByTokenRow{}, errors.New("RequireSession called without Session middleware")
	}

	slot, ok := humaCtx.Context().Value(contextSessionSlot).(*sessionSlot)
	if !ok {
		return generated.GetSessionUserByTokenRow{}, errors.New("RequireSession called without Session middleware")
	}

	if slot.filled {
		return slot.row, nil
	}

	session, plaintext, err := createNewSession(queries, humaCtx)
	if err != nil {
		slog.Error("failed to create new session", "error", err)
		return generated.GetSessionUserByTokenRow{}, err
	}

	mw.setSessionTokenCookie(humaCtx, plaintext)

	row := generated.GetSessionUserByTokenRow{
		SessionID: session.ID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	}
	slot.row = row
	slot.filled = true

	return row, nil
}

func getSessionTokenFromCookie(ctx huma.Context) string {
	cookie, err := huma.ReadCookie(ctx, cookieSessionToken)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (m *SessionMiddleware) setSessionTokenCookie(ctx huma.Context, token string) {
	newCookie := http.Cookie{
		Name:     cookieSessionToken,
		Value:    token,
		HttpOnly: true,
		Secure:   m.cfg.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(sessionTokenTTL),
	}
	ctx.AppendHeader("Set-Cookie", newCookie.String())
}

func ClearSessionTokenCookie(ctx context.Context) error {
	humaCtx, ok := humaContextFrom(ctx)
	if !ok {
		return errors.New("ClearSessionTokenCookie called without Session middleware")
	}
	mw, ok := sessionMiddlewareFrom(ctx)
	if !ok {
		return errors.New("ClearSessionTokenCookie called without Session middleware")
	}

	cookie := http.Cookie{
		Name:     cookieSessionToken,
		Value:    "",
		HttpOnly: true,
		Secure:   mw.cfg.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	}
	humaCtx.AppendHeader("Set-Cookie", cookie.String())

	return nil
}

func extendCloseToExpiredSessions(sessionUser *generated.GetSessionUserByTokenRow, queries *generated.Queries, ctx huma.Context) {
	sessionExpirationTime := sessionUser.ExpiresAt.Time
	if time.Now().Add(sessionExtensionThreshold).After(sessionExpirationTime) {
		// sessionUser.Token is already the SHA-256 hex digest stored in the
		// DB (returned by GetSessionUserByToken), so no further hashing is
		// needed before passing it back to ExtendSessionByToken.
		session, err := queries.ExtendSessionByToken(ctx.Context(), generated.ExtendSessionByTokenParams{
			Token: sessionUser.Token,
			ExpiresAt: pgtype.Timestamptz{
				Time:             time.Now().Add(sessionTokenTTL),
				InfinityModifier: pgtype.Finite,
				Valid:            true,
			},
		})
		if err == nil {
			sessionUser.ExpiresAt = session.ExpiresAt
		} else {
			slog.Warn("failed to extend new session", "error", err)
		}
	}
}

// createNewSession mints a new anonymous session. It generates a plaintext
// session token (which the caller must place in the Set-Cookie header) and
// stores the SHA-256 hex digest in the database. The plaintext is never
// persisted.
func createNewSession(queries *generated.Queries, ctx huma.Context) (generated.Session, string, error) {
	plaintext := auth.GenerateToken()
	session, err := queries.CreateSession(ctx.Context(), generated.CreateSessionParams{
		UserID: pgtype.UUID{Valid: false},
		ExpiresAt: pgtype.Timestamptz{
			Time:             time.Now().Add(sessionTokenTTL),
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
		Token: auth.HashToken(plaintext),
	})
	if err != nil {
		return generated.Session{}, "", err
	}

	return session, plaintext, nil
}

// humaContextFrom returns the huma.Context stashed by Session middleware.
func humaContextFrom(ctx context.Context) (huma.Context, bool) {
	humaCtx, ok := ctx.Value(contextHumaCtx).(huma.Context)
	return humaCtx, ok
}

// sessionMiddlewareFrom returns the SessionMiddleware stashed by the request
// pipeline.
func sessionMiddlewareFrom(ctx context.Context) (*SessionMiddleware, bool) {
	mw, ok := ctx.Value(contextSessionMW).(*SessionMiddleware)
	return mw, ok
}
