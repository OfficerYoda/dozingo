package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/avatar"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		panic(err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping database", "error", err)
		panic(err)
	}
	slog.Info("Connected to database")

	if err := seed(pool); err != nil {
		slog.Error("seeding failed", "error", err)
		panic(err)
	}

	slog.Info("Seeding completed successfully")
}

func seed(pool *pgxpool.Pool) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config for avatar seeding: %w", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is a no-op after commit

	q := generated.New(tx)

	err = truncateAll(ctx, tx)
	if err != nil {
		return err
	}

	userIDs, err := seedUsers(ctx, q)
	if err != nil {
		return err
	}

	err = seedPasswords(ctx, q, userIDs)
	if err != nil {
		return err
	}

	sessionIDs, err := seedSessions(ctx, q, userIDs)
	if err != nil {
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

	if err := seedGames(ctx, q, userIDs, sessionIDs, boardIDs, cellIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	// Avatars are generated outside the DB transaction because they require
	// external HTTP calls (DiceBear) and S3 uploads (Garage). Holding a tx
	// open across those would lock rows for the duration of network I/O.
	// Seeding hard-fails if any avatar step fails so a broken dev
	// environment surfaces immediately rather than leaving users on the
	// 'default.svg' placeholder.
	if err := seedAvatars(ctx, pool, cfg, userIDs); err != nil {
		return fmt.Errorf("seeding avatars: %w", err)
	}

	return nil
}

// truncateAll removes all data from tables in the correct order (respecting foreign keys).
func truncateAll(ctx context.Context, tx pgx.Tx) error {
	slog.Info("Truncating all tables")
	_, err := tx.Exec(ctx, "TRUNCATE game_cells, games, votes, cells, boards, sessions, user_passwords, user_authentications, users CASCADE")
	if err != nil {
		return fmt.Errorf("truncating tables: %w", err)
	}
	return nil
}

func seedUsers(ctx context.Context, q *generated.Queries) ([]pgtype.UUID, error) {
	slog.Info("Seeding users", "count", len(users))

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
	slog.Info("Seeding passwords", "count", len(passwords))

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
	slog.Info("Seeding sessions", "count", len(sessions))

	ids := make([]pgtype.UUID, 0, len(sessions))
	for _, s := range sessions {
		var userID pgtype.UUID
		if s.UserIdx >= 0 {
			userID = userIDs[s.UserIdx]
		}

		sess, err := q.CreateSession(ctx, generated.CreateSessionParams{
			UserID: userID,
			// s.Token is the plaintext value a developer pastes into the
			// session_token cookie; the DB stores its SHA-256 hex digest.
			Token: auth.HashToken(s.Token),
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
	slog.Info("Seeding boards", "count", len(boards))

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

	slog.Info("Seeded cells", "total_cells", totalCells, "board_count", len(boards))
	return cellIDsByBoard, nil
}

func seedVotes(ctx context.Context, q *generated.Queries, userIDs, boardIDs []pgtype.UUID) error {
	slog.Info("Seeding votes", "count", len(votes))

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

func seedGames(ctx context.Context, q *generated.Queries, userIDs, sessionIDs, boardIDs []pgtype.UUID, cellIDsByBoard map[int][]pgtype.UUID) error {
	slog.Info("Seeding games", "count", len(games))

	for gameIdx, g := range games {
		var playerID, sessionID pgtype.UUID
		if g.PlayerIdx >= 0 {
			playerID = userIDs[g.PlayerIdx]
		}
		if g.SessionIdx >= 0 {
			sessionID = sessionIDs[g.SessionIdx]
		}

		game, err := q.CreateGame(ctx, generated.CreateGameParams{
			PlayerID:  playerID,
			SessionID: sessionID,
			BoardID:   boardIDs[g.BoardIdx],
		})
		if err != nil {
			return fmt.Errorf("creating game %d: %w", gameIdx, err)
		}

		// Update status if not "active" (default). Authorisation is by
		// player_id when the game has one, or by session_id otherwise --
		// matches the runtime behaviour of UpdateGameStatus.
		if g.Status != "active" {
			updateParams := generated.UpdateGameStatusParams{
				Status: generated.GameStatus(g.Status),
				GameID: game.ID,
			}
			if g.PlayerIdx >= 0 {
				updateParams.PlayerID = userIDs[g.PlayerIdx]
			} else {
				updateParams.SessionID = sessionIDs[g.SessionIdx]
			}
			_, err = q.UpdateGameStatus(ctx, updateParams)
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

		gameIDs := make([]pgtype.UUID, 0, len(cells))
		cellIDs := make([]pgtype.UUID, 0, len(cells))
		contents := make([]string, 0, len(cells))
		positions := make([]int32, 0, len(cells))
		for i, gc := range cells {
			// Use the corresponding cell ID from the board if available
			var cellID pgtype.UUID
			if i < len(boardCellIDs) {
				cellID = boardCellIDs[i]
			}

			gameIDs = append(gameIDs, game.ID)
			cellIDs = append(cellIDs, cellID)
			contents = append(contents, gc.Content)
			positions = append(positions, gc.Position)
		}

		_, err = q.CreateGameCells(ctx, generated.CreateGameCellsParams{
			GameIds:   gameIDs,
			CellIds:   cellIDs,
			Contents:  contents,
			Positions: positions,
		})
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
						IsMarked:   true,
						GameCellID: row.ID,
						GameID:     game.ID,
					})
					if err != nil {
						return fmt.Errorf("marking game cell at position %d for game %d: %w", gc.Position, gameIdx, err)
					}
					break
				}
			}
		}

		slog.Info("Game cells seeded", "game_idx", gameIdx, "cell_count", len(cells), "marked_count", countMarked(cells))
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

// seedAvatars uploads a deterministic default.svg plus one DiceBear avatar
// per seeded user, then writes each new key onto the user row via
// repository.Users.SetAvatar. Hard-fails on the first error so a broken
// Garage or DiceBear surfaces immediately rather than leaving rows on the
// migration's 'default.svg' placeholder.
//
// Mirrors the shape of service.Auth.assignGeneratedAvatar without depending
// on it: the seed binary intentionally avoids dragging in service.Auth's
// session/middleware/txRunner machinery for what is a straightforward
// generate-upload-persist sequence.
func seedAvatars(
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg *config.Config,
	userIDs []pgtype.UUID,
) error {
	garage := storage.NewGarage(ctx, cfg)
	repos := repository.New(pool)

	if err := uploadDefaultAvatar(ctx, garage); err != nil {
		return fmt.Errorf("default avatar: %w", err)
	}
	slog.Info("Default avatar uploaded", "key", "default.svg")

	slog.Info("Seeding user avatars", "count", len(userIDs))
	for i, userID := range userIDs {
		if i >= len(users) {
			return fmt.Errorf("seeded user index %d exceeds users slice length %d", i, len(users))
		}
		username := users[i].Username

		key, err := generateAndUploadAvatar(ctx, garage, username)
		if err != nil {
			return fmt.Errorf("avatar for %q: %w", username, err)
		}

		if _, err := repos.Users.SetAvatar(ctx, userID, key); err != nil {
			return fmt.Errorf("setting avatar key for %q: %w", username, err)
		}
	}

	slog.Info("User avatars uploaded", "count", len(userIDs))
	return nil
}

// uploadDefaultAvatar generates a deterministic miniavs SVG (seed
// "default") and PUTs it under the canonical key "default.svg" so the
// migration's NOT NULL DEFAULT 'default.svg' resolves to a real object.
// Idempotent: re-running seed simply overwrites the same key.
func uploadDefaultAvatar(ctx context.Context, uploader storage.ObjectUploader) error {
	img, err := avatar.RandomProfilePicture("default")
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	if err := uploader.Upload(ctx, "default.svg", img); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	return nil
}

// generateAndUploadAvatar fetches a DiceBear avatar for the given seed,
// uploads it under a fresh UUID-based object key, and returns the key the
// caller should persist to users.avatar_key.
func generateAndUploadAvatar(ctx context.Context, uploader storage.ObjectUploader, seed string) (string, error) {
	img, err := avatar.RandomProfilePicture(seed)
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("uuid: %w", err)
	}

	key := fmt.Sprintf("%s%s", id, img.Extension)
	if err := uploader.Upload(ctx, key, img); err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}
	return key, nil
}
