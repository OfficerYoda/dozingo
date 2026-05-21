package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

// Repos is the bundle of all repositories wired to a single DBTX
type Repos struct {
	Boards    *Boards
	Cells     *Cells
	GameCells *GameCells
	Games     *Games
	Passwords *UserPasswords
	Sessions  *Sessions
	Users     *Users
	Votes     *Votes
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
		Cells:     &Cells{queries: queries},
		GameCells: &GameCells{queries: queries},
		Games:     &Games{queries: queries},
		Passwords: &UserPasswords{queries: queries},
		Sessions:  &Sessions{queries: queries},
		Users:     &Users{queries: queries},
		Votes:     &Votes{queries: queries},
	}
}
