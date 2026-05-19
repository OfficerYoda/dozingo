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
	"github.com/officeryoda/dozingo/internal/generated"
)

type contextKey int

const (
	contextSessionUser contextKey = iota
	contextSessionSlot
	contextHumaCtx
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

// SessionUser is a read-only middleware. If the request carries a valid
// session_token cookie, it loads the session, optionally extends it, and
// stashes it in the request context. Requests without a cookie do not touch
// the database. Handlers that need to persist anonymous activity should call
// RequireSession to lazily mint a session.
func SessionUser(api huma.API, queries *generated.Queries) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Always attach an empty slot so RequireSession can cache a minted
		// session within the lifetime of this request.
		slot := &sessionSlot{}
		ctx = huma.WithValue(ctx, contextSessionSlot, slot)

		// Stash the huma.Context itself so typed handlers (which only see
		// context.Context) can reach AppendHeader/ReadCookie via the
		// RequireSessionCtx helper.
		ctx = huma.WithValue(ctx, contextHumaCtx, ctx)

		sessionToken := getSessionTokenFromCookie(ctx)
		if sessionToken == "" {
			next(ctx)
			return
		}

		sessionUser, err := queries.GetSessionUserByToken(ctx.Context(), sessionToken)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				clearSessionTokenCookie(ctx)
				next(ctx)
				return
			}
			slog.Error("failed to look up session", "error", err)
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error")
			return
		}

		extendCloseToExpiredSessions(&sessionUser, queries, ctx)

		slot.row = sessionUser
		slot.filled = true
		ctx = huma.WithValue(ctx, contextSessionUser, sessionUser)
		next(ctx)
	}
}

// humaContextFrom returns the huma.Context stashed by SessionUser middleware.
func humaContextFrom(ctx context.Context) (huma.Context, bool) {
	humaCtx, ok := ctx.Value(contextHumaCtx).(huma.Context)
	return humaCtx, ok
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
func RequireSession(ctx huma.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	slot, ok := ctx.Context().Value(contextSessionSlot).(*sessionSlot)
	if !ok {
		return generated.GetSessionUserByTokenRow{}, errors.New("RequireSession called without SessionUser middleware")
	}

	if slot.filled {
		return slot.row, nil
	}

	session, err := createNewSession(queries, ctx)
	if err != nil {
		slog.Error("failed to create new session", "error", err)
		return generated.GetSessionUserByTokenRow{}, err
	}

	setSessionTokenCookie(ctx, session.Token)

	row := generated.GetSessionUserByTokenRow{
		SessionID: session.ID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	}
	slot.row = row
	slot.filled = true

	return row, nil
}

// RequireSessionCtx is the context.Context-friendly variant of RequireSession,
// for use inside typed Huma handlers whose signature is
// `func(ctx context.Context, input *T) (*U, error)`. It looks up the
// huma.Context that SessionUser middleware stashed and delegates.
func RequireSessionCtx(ctx context.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	humaCtx, ok := humaContextFrom(ctx)
	if !ok {
		return generated.GetSessionUserByTokenRow{}, errors.New("RequireSessionCtx called without SessionUser middleware")
	}
	return RequireSession(humaCtx, queries)
}

func getSessionTokenFromCookie(ctx huma.Context) string {
	cookie, err := huma.ReadCookie(ctx, cookieSessionToken)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setSessionTokenCookie(ctx huma.Context, token string) {
	newCookie := http.Cookie{
		Name:     cookieSessionToken,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(sessionTokenTTL),
	}
	ctx.AppendHeader("Set-Cookie", newCookie.String())
}

func clearSessionTokenCookie(ctx huma.Context) {
	cookie := http.Cookie{
		Name:     cookieSessionToken,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	}
	ctx.AppendHeader("Set-Cookie", cookie.String())
}

func extendCloseToExpiredSessions(sessionUser *generated.GetSessionUserByTokenRow, queries *generated.Queries, ctx huma.Context) {
	sessionExpirationTime := sessionUser.ExpiresAt.Time
	if time.Now().Add(sessionExtensionThreshold).After(sessionExpirationTime) {
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

func createNewSession(queries *generated.Queries, ctx huma.Context) (generated.Session, error) {
	session, err := queries.CreateSession(ctx.Context(), generated.CreateSessionParams{
		UserID: pgtype.UUID{Valid: false},
		ExpiresAt: pgtype.Timestamptz{
			Time:             time.Now().Add(sessionTokenTTL),
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
		Token: auth.GenerateSessionToken(),
	})
	if err != nil {
		return generated.Session{}, err
	}

	return session, nil
}
