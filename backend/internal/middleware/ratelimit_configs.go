package middleware

import (
	"time"

	"github.com/go-chi/httprate"
)

var (
	// EXAMPLE: Auth look at handler/auth.go for implementation
	// NOTE: I dont like the naming strict and heavy because i feel like both could be the stricter one
	StrictAuthLimiter = httprate.NewRateLimiter(5, 1*time.Minute)
	HeavyAuthLimiter  = httprate.NewRateLimiter(3, 1*time.Minute)
)
