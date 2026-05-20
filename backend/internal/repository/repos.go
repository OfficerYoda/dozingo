package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

// Repos is the bundle of all repositories wired to a single DBTX
type Repos struct {
	Boards    *Boards
	Users     *Users
	Passwords *UserPasswords
	Sessions  *Sessions
}

// New returns a Repos bundle backed by the given pool.
func New(pool *pgxpool.Pool) Repos {
	return NewWithDBTX(pool)
}

// NewWithDBTX returns a Repos bundle backed by an arbitrary DBTX.
func NewWithDBTX(db generated.DBTX) Repos {
	queries := generated.New(db)
	return Repos{
		Boards:    &Boards{queries: queries, db: db},
		Users:     &Users{queries: queries},
		Passwords: &UserPasswords{queries: queries},
		Sessions:  &Sessions{queries: queries},
	}
}
