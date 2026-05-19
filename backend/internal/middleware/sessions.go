package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

type contextKey int

const (
	cookieSessionToken                   = "session_token"
	contextSessionUser        contextKey = iota
	sessionTokenTTL                      = 30 * 24 * time.Hour
	sessionExtensionThreshold            = 7 * 24 * time.Hour
)

func SessionUser(api huma.API, queries *generated.Queries) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		sessionToken := getSessionTokenFromCookie(ctx)

		sessionUser, err := getValidSessionUser(queries, ctx, sessionToken)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error")
			return
		}

		extendCloseToExpiredSessions(&sessionUser, queries, ctx)

		ctx = huma.WithValue(ctx, contextSessionUser, sessionUser)
		next(ctx)
	}
}

func getSessionTokenFromCookie(ctx huma.Context) string {
	cookie, err := huma.ReadCookie(ctx, cookieSessionToken)
	sessionToken := ""
	if err == nil {
		sessionToken = cookie.Value
	} else {
		slog.Info("no session token cookie")
	}
	return sessionToken
}

func getValidSessionUser(queries *generated.Queries, ctx huma.Context, sessionToken string) (generated.GetSessionUserByTokenRow, error) {
	if sessionToken != "" {
		sessionUser, err := queries.GetSessionUserByToken(ctx.Context(), sessionToken)
		if err == nil {
			return sessionUser, nil
		}
		slog.Info("couldn't get sessionUser from token", "error", err)
	}

	session, err := createNewSession(queries, ctx)
	if err != nil {
		slog.Error("failed to create new session", "error", err)
		return generated.GetSessionUserByTokenRow{}, err
	}

	setSessionTokenCookie(session.Token, ctx)

	return generated.GetSessionUserByTokenRow{
		SessionID: session.ID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func setSessionTokenCookie(token string, ctx huma.Context) {
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
