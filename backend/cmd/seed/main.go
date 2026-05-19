package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/generated"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	if err := seed(pool); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully")
}

func seed(pool *pgxpool.Pool) error {
	ctx := context.Background()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is a no-op after commit

	q := generated.New(tx)

	if err := truncateAll(ctx, tx); err != nil {
		return err
	}

	userIDs, err := seedUsers(ctx, q)
	if err != nil {
		return err
	}

	if err := seedPasswords(ctx, q, userIDs); err != nil {
		return err
	}

	if _, err := seedSessions(ctx, q, userIDs); err != nil {
		return err
	}

	boardIDs, err := seedBoards(ctx, q, userIDs)
	if err != nil {
		return err
	}

	cellIDs, err := seedCells(ctx, q, boardIDs)
	if err != nil {
		return err
	}

	if err := seedVotes(ctx, q, userIDs, boardIDs); err != nil {
		return err
	}

	if err := seedGames(ctx, q, userIDs, boardIDs, cellIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// truncateAll removes all data from tables in the correct order (respecting foreign keys).
func truncateAll(ctx context.Context, tx pgx.Tx) error {
	log.Println("Truncating all tables...")
	_, err := tx.Exec(ctx, "TRUNCATE game_cells, games, votes, cells, boards, sessions, user_passwords, user_authentications, users CASCADE")
	if err != nil {
		return fmt.Errorf("truncating tables: %w", err)
	}
	return nil
}

func seedUsers(ctx context.Context, q *generated.Queries) ([]pgtype.UUID, error) {
	log.Printf("Seeding %d users...", len(users))

	ids := make([]pgtype.UUID, 0, len(users))
	for _, u := range users {

		var email pgtype.Text
		if strings.TrimSpace(u.Email) != "" {
			email = pgtype.Text{String: u.Email, Valid: true}
		} else {
			email = pgtype.Text{String: "", Valid: false}
		}
		user, err := q.CreateUser(ctx, generated.CreateUserParams{
			Username: u.Username,
			Email:    email,
		})
		if err != nil {
			return nil, fmt.Errorf("creating user %q: %w", u.Username, err)
		}
		ids = append(ids, user.ID)
	}

	return ids, nil
}

func seedPasswords(ctx context.Context, q *generated.Queries, userIDs []pgtype.UUID) error {
	log.Printf("Seeding %d passwords...", len(passwords))

	for _, p := range passwords {
		hash, err := auth.HashPassword(p.Password)
		if err != nil {
			return fmt.Errorf("hashing password for user %d: %w", p.UserIdx, err)
		}

		_, err = q.UpsertUserPassword(ctx, generated.UpsertUserPasswordParams{
			UserID:       userIDs[p.UserIdx],
			PasswordHash: hash,
		})
		if err != nil {
			return fmt.Errorf("creating password for user %d: %w", p.UserIdx, err)
		}
	}

	return nil
}

// seedSessions inserts the predefined `sessions` rows. Anonymous sessions
// (UserIdx == -1) are stored with a NULL user_id. Returns the inserted session
// IDs in the order of the `sessions` slice so callers can reference them later
// (e.g. for seeding anonymous games).
func seedSessions(ctx context.Context, q *generated.Queries, userIDs []pgtype.UUID) ([]pgtype.UUID, error) {
	log.Printf("Seeding %d sessions...", len(sessions))

	ids := make([]pgtype.UUID, 0, len(sessions))
	for _, s := range sessions {
		var userID pgtype.UUID
		if s.UserIdx >= 0 {
			userID = userIDs[s.UserIdx]
		}

		sess, err := q.CreateSession(ctx, generated.CreateSessionParams{
			UserID: userID,
			Token:  s.Token,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(time.Duration(s.ExpiresInHours) * time.Hour),
				Valid: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("creating session %q: %w", s.Token, err)
		}
		ids = append(ids, sess.ID)
	}

	return ids, nil
}

func seedBoards(ctx context.Context, q *generated.Queries, userIDs []pgtype.UUID) ([]pgtype.UUID, error) {
	log.Printf("Seeding %d boards...", len(boards))

	ids := make([]pgtype.UUID, 0, len(boards))
	for _, b := range boards {
		board, err := q.CreateBoard(ctx, generated.CreateBoardParams{
			Title:       b.Title,
			Description: pgtype.Text{String: b.Description, Valid: b.Description != ""},
			Size:        b.Size,
			AuthorID:    userIDs[b.AuthorIdx],
		})
		if err != nil {
			return nil, fmt.Errorf("creating board %q: %w", b.Title, err)
		}
		ids = append(ids, board.ID)
	}

	return ids, nil
}

// seedCells creates cells for each board and returns a map of boardIdx -> []cellID.
func seedCells(ctx context.Context, q *generated.Queries, boardIDs []pgtype.UUID) (map[int][]pgtype.UUID, error) {
	cellIDsByBoard := make(map[int][]pgtype.UUID)
	totalCells := 0

	for i, b := range boards {
		phrases, ok := cellPhrases[i]
		if !ok {
			return nil, fmt.Errorf("no cell phrases defined for board %d (%q)", i, b.Title)
		}

		cellIDs := make([]pgtype.UUID, 0, len(phrases))
		for _, phrase := range phrases {
			cell, err := q.CreateCell(ctx, generated.CreateCellParams{
				BoardID: boardIDs[i],
				Content: phrase,
				Value:   1,
			})
			if err != nil {
				return nil, fmt.Errorf("creating cell for board %q: %w", b.Title, err)
			}
			cellIDs = append(cellIDs, cell.ID)
			totalCells++
		}
		cellIDsByBoard[i] = cellIDs
	}

	log.Printf("Seeded %d cells across %d boards", totalCells, len(boards))
	return cellIDsByBoard, nil
}

func seedVotes(ctx context.Context, q *generated.Queries, userIDs, boardIDs []pgtype.UUID) error {
	log.Printf("Seeding %d votes...", len(votes))

	for _, v := range votes {
		_, err := q.UpsertVote(ctx, generated.UpsertVoteParams{
			UserID:    userIDs[v.UserIdx],
			BoardID:   boardIDs[v.BoardIdx],
			VoteValue: v.Value,
		})
		if err != nil {
			return fmt.Errorf("creating vote (user=%d, board=%d): %w", v.UserIdx, v.BoardIdx, err)
		}
	}

	return nil
}

func seedGames(ctx context.Context, q *generated.Queries, userIDs, boardIDs []pgtype.UUID, cellIDsByBoard map[int][]pgtype.UUID) error {
	log.Printf("Seeding %d games...", len(games))

	for gameIdx, g := range games {
		game, err := q.CreateGame(ctx, generated.CreateGameParams{
			PlayerID: userIDs[g.PlayerIdx],
			BoardID:  boardIDs[g.BoardIdx],
		})
		if err != nil {
			return fmt.Errorf("creating game %d: %w", gameIdx, err)
		}

		// Update status if not "active" (default)
		if g.Status != "active" {
			_, err = q.UpdateGameStatus(ctx, generated.UpdateGameStatusParams{
				Status:   g.Status,
				ID:       game.ID,
				PlayerID: userIDs[g.PlayerIdx],
			})
			if err != nil {
				return fmt.Errorf("updating game %d status to %q: %w", gameIdx, g.Status, err)
			}
		}

		// Seed game cells
		cells, ok := gameCells[gameIdx]
		if !ok {
			continue
		}

		boardCellIDs := cellIDsByBoard[g.BoardIdx]

		params := make([]generated.CreateGameCellsParams, 0, len(cells))
		for i, gc := range cells {
			// Use the corresponding cell ID from the board if available
			var cellID pgtype.UUID
			if i < len(boardCellIDs) {
				cellID = boardCellIDs[i]
			}

			params = append(params, generated.CreateGameCellsParams{
				GameID:   game.ID,
				CellID:   cellID,
				Content:  gc.Content,
				Position: gc.Position,
			})
		}

		_, err = q.CreateGameCells(ctx, params)
		if err != nil {
			return fmt.Errorf("creating game cells for game %d: %w", gameIdx, err)
		}

		// Mark cells that should be marked
		// We need to fetch them first since CopyFrom doesn't return IDs
		gameCellRows, err := q.GetGameCellsByGameID(ctx, game.ID)
		if err != nil {
			return fmt.Errorf("fetching game cells for game %d: %w", gameIdx, err)
		}

		for _, row := range gameCellRows {
			// Find matching cell data by position
			for _, gc := range cells {
				if gc.Position == row.Position && gc.IsMarked {
					_, err = q.UpdateGameCellMark(ctx, generated.UpdateGameCellMarkParams{
						IsMarked: true,
						ID:       row.ID,
						GameID:   game.ID,
					})
					if err != nil {
						return fmt.Errorf("marking game cell at position %d for game %d: %w", gc.Position, gameIdx, err)
					}
					break
				}
			}
		}

		log.Printf("  Game %d: %d cells seeded (%d marked)", gameIdx, len(cells), countMarked(cells))
	}

	return nil
}

func countMarked(cells []gameCellData) int {
	count := 0
	for _, c := range cells {
		if c.IsMarked {
			count++
		}
	}
	return count
}
