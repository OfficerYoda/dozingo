package middleware

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/httprate"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

func RateLimit(
	api huma.API,
	rl *httprate.RateLimiter,
) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		r, w := humachi.Unwrap(ctx)

		key, err := sessKeyFn(r)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "rate limit key error")
			return
		}

		// Combine session/IP with the endpoint so limits are per-route.
		key = key + ":" + ctx.Operation().Path

		if rl.OnLimit(w, r, key) {
			// OnLimit already wrote 429 + headers; stop the chain.
			return
		}

		next(ctx)
	}
}

func sessKeyFn(r *http.Request) (string, error) {
	sessionUser, ok := r.Context().Value(contextSessionUser).(generated.GetSessionUserByTokenRow)
	if ok && sessionUser.SessionID.Valid {
		return "sess:" + *pgmap.StringFromPgUUID(sessionUser.SessionID), nil
	}

	return ipKeyFn(r)
}

func ipKeyFn(r *http.Request) (string, error) {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i != -1 {
			return strings.TrimSpace(xff[:i]), nil
		}
		return strings.TrimSpace(xff), nil
	}
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	return ip, nil
}
