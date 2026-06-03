package middleware

import (
	"time"

	"github.com/go-chi/httprate"
)

var (
	// Auth
	StrictAuthLimiter = httprate.NewRateLimiter(5, 1*time.Minute) // register, forgot/new password, email verification flows
	HeavyAuthLimiter  = httprate.NewRateLimiter(3, 1*time.Minute) // login

	// Resource mutations
	WriteLimiter      = httprate.NewRateLimiter(60, 1*time.Minute) // typical creates/updates/deletes
	WriteHeavyLimiter = httprate.NewRateLimiter(20, 1*time.Minute) // bulk write endpoints (POST /games/{id}/cells)

	// Reads
	ReadLimiter     = httprate.NewRateLimiter(120, 1*time.Minute) // single-resource & own-resource GETs
	ReadListLimiter = httprate.NewRateLimiter(60, 1*time.Minute)  // search/list endpoints

	// Gameplay (high-frequency mark toggles)
	GameplayLimiter = httprate.NewRateLimiter(120, 1*time.Minute)

	// Public/infra (typically IP-keyed since usually unauthenticated)
	HealthLimiter = httprate.NewRateLimiter(30, 1*time.Minute)
)
