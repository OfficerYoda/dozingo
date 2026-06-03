package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/generated"
)

// ===== Helpers =====

// recordingHandler is a no-op huma operation handler that increments a
// counter on every invocation. Tests use it to verify that the rate
// limiter actually short-circuits the chain (i.e. handler.calls == limit
// even after limit+1 requests).
type recordingHandler struct {
	calls atomic.Int32
}

func (h *recordingHandler) op(_ context.Context, _ *struct{}) (*struct{}, error) {
	h.calls.Add(1)
	return &struct{}{}, nil
}

// rlTestOp describes a single operation to register on the test API.
type rlTestOp struct {
	path    string
	limiter *httprate.RateLimiter
}

// newRateLimitTestRouter builds an isolated chi+huma app, registers each op
// behind RateLimit middleware, and returns the chi mux and the per-op
// recordingHandlers in the same order as ops.
func newRateLimitTestRouter(t *testing.T, ops ...rlTestOp) (http.Handler, []*recordingHandler) {
	t.Helper()
	r := chi.NewMux()
	api := humachi.New(r, huma.DefaultConfig("rl-test", "0.0.1"))

	handlers := make([]*recordingHandler, len(ops))
	for i, op := range ops {
		h := &recordingHandler{}
		handlers[i] = h
		huma.Register(api, huma.Operation{
			OperationID: "op-" + op.path,
			Method:      http.MethodGet,
			Path:        op.path,
			Middlewares: huma.Middlewares{RateLimit(api, op.limiter)},
		}, h.op)
	}
	return r, handlers
}

// reqWithSession returns a GET request to path with the given session-user
// stashed in the request context (mimicking what SessionMiddleware does in
// production). Pass an empty pgtype.UUID to simulate a session-user with no
// valid SessionID.
func reqWithSession(path string, sessionID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	row := generated.GetSessionUserByTokenRow{SessionID: sessionID}
	ctx := context.WithValue(req.Context(), contextSessionUser, row)
	return req.WithContext(ctx)
}

// reqWithRemoteAddr returns a GET request to path with the given RemoteAddr.
// No session-user is set, so the rate limiter falls back to IP keying.
func reqWithRemoteAddr(path, remoteAddr string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = remoteAddr
	return req
}

// pgUUID returns a pgtype.UUID with Valid=true backed by a fresh random UUID.
func pgUUID(t *testing.T) pgtype.UUID {
	t.Helper()
	id := uuid.New()
	var u pgtype.UUID
	copy(u.Bytes[:], id[:])
	u.Valid = true
	return u
}

// ===== ipKeyFn tests =====

func TestIPKeyFn(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{
			name:       "RemoteAddr with port",
			remoteAddr: "1.2.3.4:5678",
			want:       "1.2.3.4",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "1.2.3.4",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:1234",
			xff:        "5.6.7.8",
			want:       "5.6.7.8",
		},
		{
			name:       "X-Forwarded-For comma list takes first",
			remoteAddr: "10.0.0.1:1234",
			xff:        "5.6.7.8, 9.10.11.12",
			want:       "5.6.7.8",
		},
		{
			name:       "X-Forwarded-For trims whitespace",
			remoteAddr: "10.0.0.1:1234",
			xff:        "  5.6.7.8  ,9.10.11.12",
			want:       "5.6.7.8",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			got, err := ipKeyFn(req)
			if err != nil {
				t.Fatalf("ipKeyFn returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("ipKeyFn = %q, want %q", got, tc.want)
			}
		})
	}
}

// ===== sessKeyFn tests =====

func TestSessKeyFn_WithSession(t *testing.T) {
	sid := pgUUID(t)
	req := reqWithSession("/x", sid)

	got, err := sessKeyFn(req)
	if err != nil {
		t.Fatalf("sessKeyFn returned unexpected error: %v", err)
	}

	want := "sess:" + uuid.UUID(sid.Bytes).String()
	if got != want {
		t.Errorf("sessKeyFn = %q, want %q", got, want)
	}
}

func TestSessKeyFn_FallsBackToIP_WhenNoSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "9.9.9.9:1111"

	got, err := sessKeyFn(req)
	if err != nil {
		t.Fatalf("sessKeyFn returned unexpected error: %v", err)
	}
	if got != "9.9.9.9" {
		t.Errorf("sessKeyFn fallback = %q, want %q", got, "9.9.9.9")
	}
}

func TestSessKeyFn_FallsBackToIP_WhenSessionUUIDInvalid(t *testing.T) {
	// Session-user is in context but its SessionID has Valid=false.
	req := reqWithSession("/x", pgtype.UUID{Valid: false})
	req.RemoteAddr = "9.9.9.9:1111"

	got, err := sessKeyFn(req)
	if err != nil {
		t.Fatalf("sessKeyFn returned unexpected error: %v", err)
	}
	if got != "9.9.9.9" {
		t.Errorf("sessKeyFn fallback (invalid SessionID) = %q, want %q", got, "9.9.9.9")
	}
}

// ===== RateLimit middleware tests =====

// successStatus is the status code huma uses when a handler returns
// (*struct{}, nil): no body, so 204 No Content.
const successStatus = http.StatusNoContent

func TestRateLimit_AllowsRequestsUnderLimit(t *testing.T) {
	rl := httprate.NewRateLimiter(2, time.Minute)
	router, handlers := newRateLimitTestRouter(t, rlTestOp{path: "/op", limiter: rl})
	h := handlers[0]

	sid := pgUUID(t)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithSession("/op", sid))
		if w.Code != successStatus {
			t.Fatalf("request %d: expected %d, got %d (body=%s)", i+1, successStatus, w.Code, w.Body.String())
		}
	}
	if got := h.calls.Load(); got != 2 {
		t.Errorf("handler calls = %d, want 2", got)
	}
}

func TestRateLimit_BlocksRequestsOverLimit(t *testing.T) {
	rl := httprate.NewRateLimiter(2, time.Minute)
	router, handlers := newRateLimitTestRouter(t, rlTestOp{path: "/op", limiter: rl})
	h := handlers[0]

	sid := pgUUID(t)

	// First 2 requests succeed.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithSession("/op", sid))
		if w.Code != successStatus {
			t.Fatalf("request %d: expected %d, got %d", i+1, successStatus, w.Code)
		}
	}

	// 3rd request must be blocked with 429.
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqWithSession("/op", sid))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request: expected 429, got %d (body=%s)", w.Code, w.Body.String())
	}

	// Handler must have been short-circuited on the 3rd request.
	if got := h.calls.Load(); got != 2 {
		t.Errorf("handler calls = %d, want 2 (limiter did not short-circuit)", got)
	}
}

func TestRateLimit_PerSessionIsolation(t *testing.T) {
	rl := httprate.NewRateLimiter(2, time.Minute)
	router, handlers := newRateLimitTestRouter(t, rlTestOp{path: "/op", limiter: rl})
	h := handlers[0]

	sidA := pgUUID(t)
	sidB := pgUUID(t)

	// Session A: exhaust quota.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithSession("/op", sidA))
		if w.Code != successStatus {
			t.Fatalf("session A request %d: expected %d, got %d", i+1, successStatus, w.Code)
		}
	}

	// Session A: now over limit.
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqWithSession("/op", sidA))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("session A 3rd request: expected 429, got %d", w.Code)
	}

	// Session B: must still have a fresh quota.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithSession("/op", sidB))
		if w.Code != successStatus {
			t.Errorf("session B request %d: expected %d (independent counter), got %d", i+1, successStatus, w.Code)
		}
	}

	// Total successful invocations: 2 (A) + 2 (B) = 4.
	if got := h.calls.Load(); got != 4 {
		t.Errorf("handler calls = %d, want 4", got)
	}
}

func TestRateLimit_PerPathIsolation(t *testing.T) {
	// SAME limiter struct registered on two different ops/paths; the
	// per-route key suffix in RateLimit must keep their counters separate.
	rl := httprate.NewRateLimiter(2, time.Minute)
	router, handlers := newRateLimitTestRouter(t,
		rlTestOp{path: "/a", limiter: rl},
		rlTestOp{path: "/b", limiter: rl},
	)
	hA, hB := handlers[0], handlers[1]

	sid := pgUUID(t)

	// Exhaust /a.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithSession("/a", sid))
		if w.Code != successStatus {
			t.Fatalf("/a request %d: expected %d, got %d", i+1, successStatus, w.Code)
		}
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqWithSession("/a", sid))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("/a 3rd request: expected 429, got %d", w.Code)
	}

	// /b must still have a fresh quota despite sharing the limiter struct.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithSession("/b", sid))
		if w.Code != successStatus {
			t.Errorf("/b request %d: expected %d (per-path isolation), got %d", i+1, successStatus, w.Code)
		}
	}

	if got := hA.calls.Load(); got != 2 {
		t.Errorf("/a handler calls = %d, want 2", got)
	}
	if got := hB.calls.Load(); got != 2 {
		t.Errorf("/b handler calls = %d, want 2", got)
	}
}

func TestRateLimit_IPFallbackIsolation(t *testing.T) {
	rl := httprate.NewRateLimiter(2, time.Minute)
	router, handlers := newRateLimitTestRouter(t, rlTestOp{path: "/op", limiter: rl})
	h := handlers[0]

	// IP A: exhaust quota.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithRemoteAddr("/op", "1.1.1.1:1234"))
		if w.Code != successStatus {
			t.Fatalf("IP A request %d: expected %d, got %d", i+1, successStatus, w.Code)
		}
	}

	// IP A over limit.
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqWithRemoteAddr("/op", "1.1.1.1:1234"))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP A 3rd request: expected 429, got %d", w.Code)
	}

	// IP B must have a fresh quota.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, reqWithRemoteAddr("/op", "2.2.2.2:5678"))
		if w.Code != successStatus {
			t.Errorf("IP B request %d: expected %d (independent counter), got %d", i+1, successStatus, w.Code)
		}
	}

	if got := h.calls.Load(); got != 4 {
		t.Errorf("handler calls = %d, want 4", got)
	}
}
