package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/domain"
)

// statusAndBody extracts the HTTP status code and response detail string
// from an error returned by toHumaErr. It fails the test if err does not
// satisfy huma.StatusError.
func statusAndBody(t *testing.T, err error) (int, string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var se huma.StatusError
	if !errors.As(err, &se) {
		t.Fatalf("expected huma.StatusError, got %T (%v)", err, err)
	}
	return se.GetStatus(), se.Error()
}

// assertNoLeak fails the test if body contains any of the given substrings.
// Used to verify that internal details (constraint names, wrap context)
// are not present in client-facing error messages.
func assertNoLeak(t *testing.T, body string, leaks ...string) {
	t.Helper()
	for _, leak := range leaks {
		if strings.Contains(body, leak) {
			t.Errorf("response body %q must not contain leaked substring %q", body, leak)
		}
	}
}

func TestToHumaErr_NotFound_DefaultMessage(t *testing.T) {
	got := toHumaErr(domain.ErrNotFound, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
	if body != msgNotFound {
		t.Errorf("expected body %q, got %q", msgNotFound, body)
	}
}

func TestToHumaErr_NotFound_OverrideMessage(t *testing.T) {
	got := toHumaErr(domain.ErrNotFound, "board not found", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
	if body != "board not found" {
		t.Errorf("expected override body %q, got %q", "board not found", body)
	}
}

func TestToHumaErr_NotFound_DoesNotLeakWrappedDetails(t *testing.T) {
	wrapped := fmt.Errorf("internal repo context: %w", domain.ErrNotFound)
	got := toHumaErr(wrapped, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
	if body != msgNotFound {
		t.Errorf("expected body %q, got %q", msgNotFound, body)
	}
	assertNoLeak(t, body, "internal repo context")
}

func TestToHumaErr_Conflict_DoesNotLeakConstraint(t *testing.T) {
	// Mimic pgmap.TranslatePgErr's wrapping for a unique_violation.
	wrapped := fmt.Errorf("users_email_key: %w", domain.ErrConflict)
	got := toHumaErr(wrapped, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusConflict {
		t.Errorf("expected status 409, got %d", status)
	}
	if body != msgConflict {
		t.Errorf("expected body %q, got %q", msgConflict, body)
	}
	assertNoLeak(t, body, "users_email_key", "_key")
}

func TestToHumaErr_BadInput_DoesNotLeakConstraint(t *testing.T) {
	// Mimic pgmap.TranslatePgErr's wrapping for a check_violation.
	wrapped := fmt.Errorf("votes_vote_value_check: %w", domain.ErrBadInput)
	got := toHumaErr(wrapped, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", status)
	}
	if body != msgBadRequest {
		t.Errorf("expected body %q, got %q", msgBadRequest, body)
	}
	assertNoLeak(t, body, "votes_vote_value_check", "_check")
}

func TestToHumaErr_Unauthorized_DoesNotLeakWrap(t *testing.T) {
	wrapped := fmt.Errorf("session required: %w", domain.ErrUnauthorized)
	got := toHumaErr(wrapped, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", status)
	}
	if body != msgUnauthorized {
		t.Errorf("expected body %q, got %q", msgUnauthorized, body)
	}
	assertNoLeak(t, body, "session required")
}

func TestToHumaErr_Forbidden_DoesNotLeakWrap(t *testing.T) {
	wrapped := fmt.Errorf("caller doesn't own board: %w", domain.ErrForbidden)
	got := toHumaErr(wrapped, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
	if body != msgForbidden {
		t.Errorf("expected body %q, got %q", msgForbidden, body)
	}
	assertNoLeak(t, body, "caller doesn't own")
}

func TestToHumaErr_UnprocessableEntity_DoesNotLeakWrap(t *testing.T) {
	wrapped := fmt.Errorf("invalid vote_value: %w", domain.ErrUnprocessableEntity)
	got := toHumaErr(wrapped, "", "op")
	status, body := statusAndBody(t, got)
	if status != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", status)
	}
	if body != msgUnprocessableEntity {
		t.Errorf("expected body %q, got %q", msgUnprocessableEntity, body)
	}
	assertNoLeak(t, body, "invalid vote_value")
}

func TestToHumaErr_UnmatchedReturns500WithOpMsg(t *testing.T) {
	raw := errors.New("raw db connection blew up")
	got := toHumaErr(raw, "", "failed to do thing")
	status, body := statusAndBody(t, got)
	if status != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", status)
	}
	if body != "failed to do thing" {
		t.Errorf("expected body %q, got %q", "failed to do thing", body)
	}
	assertNoLeak(t, body, "raw db connection")
}

func TestToHumaErr_TwoFARequired_Returns403(t *testing.T) {
	wrapped := fmt.Errorf("2fa required: %w", domain.ErrTwoFARequired)
	got := toHumaErr(wrapped, "", "failed to login user")
	status, body := statusAndBody(t, got)
	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
	if body != msgTwoFARequired {
		t.Errorf("expected body %q, got %q", msgTwoFARequired, body)
	}
	assertNoLeak(t, body, "2fa required", "failed to login user")
}
