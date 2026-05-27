package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Boards struct {
	queries *generated.Queries
	db      generated.DBTX
}

type BoardListFilter struct {
	AuthorID string
	Size     int32
	Sort     string
	Limit    int32
}

type CreateBoardInput struct {
	Title       string
	Description *string
	Size        int32
	AuthorID    pgtype.UUID
}

// BoardWithStats is a board augmented with aggregate counters: total vote
// score (sum of vote_value), number of votes cast, and number of games
// played. Returned by all repository read paths.
type BoardWithStats struct {
	generated.Board
	Score     int64 `db:"score"`
	VoteCount int64 `db:"vote_count"`
	PlayCount int64 `db:"play_count"`
}

// boardWithStatsBaseSelect is the SELECT clause + FROM/JOIN tree shared by
// list and get-by-id queries. Both join pre-aggregated subqueries so
// per-board counts stay accurate even when filters or LIMIT are applied,
// without inflation from a Cartesian product between votes and games.
const boardWithStatsBaseSelect = `SELECT
	b.id, b.title, b.size, b.author_id, b.created_at, b.updated_at, b.description,
	COALESCE(v.score, 0)      AS score,
	COALESCE(v.vote_count, 0) AS vote_count,
	COALESCE(g.play_count, 0) AS play_count
FROM boards b
LEFT JOIN (
	SELECT board_id,
	       SUM(vote_value) AS score,
	       COUNT(*)        AS vote_count
	FROM votes
	GROUP BY board_id
) v ON v.board_id = b.id
LEFT JOIN (
	SELECT board_id, COUNT(*) AS play_count
	FROM games
	GROUP BY board_id
) g ON g.board_id = b.id`

// orderByForSort returns the ORDER BY expression (without the "ORDER BY"
// prefix) for the given sort key. Unknown / empty values fall back to
// newest-first. Non-temporal sorts include a deterministic created_at
// tiebreaker.
func orderByForSort(sort string) string {
	switch sort {
	case "oldest":
		return "b.created_at ASC"
	case "most-liked":
		return "COALESCE(v.score, 0) DESC, b.created_at DESC"
	case "least-liked":
		return "COALESCE(v.score, 0) ASC, b.created_at DESC"
	case "most-played":
		return "COALESCE(g.play_count, 0) DESC, b.created_at DESC"
	case "least-played":
		return "COALESCE(g.play_count, 0) ASC, b.created_at DESC"
	default:
		return "b.created_at DESC"
	}
}

func (r *Boards) List(ctx context.Context, f BoardListFilter) ([]BoardWithStats, error) {
	var (
		query strings.Builder
		args  []any
		i     = 1
	)

	query.WriteString(boardWithStatsBaseSelect)
	query.WriteString(" WHERE 1=1")

	if f.AuthorID != "" {
		var authorUUID pgtype.UUID
		err := authorUUID.Scan(f.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("invalid author_id: %w", domain.ErrBadInput)
		}
		fmt.Fprintf(&query, " AND b.author_id = $%d", i)
		args = append(args, authorUUID)
		i++
	}

	if f.Size != 0 {
		fmt.Fprintf(&query, " AND b.size = $%d", i)
		args = append(args, f.Size)
		i++
	}

	query.WriteString(" ORDER BY ")
	query.WriteString(orderByForSort(f.Sort))

	if f.Limit > 0 {
		fmt.Fprintf(&query, " LIMIT $%d", i)
		args = append(args, f.Limit)
		i++
	}

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query boards: %w", err)
	}
	defer rows.Close()

	boards, err := pgx.CollectRows(rows, pgx.RowToStructByName[BoardWithStats])
	if err != nil {
		return nil, fmt.Errorf("scan boards: %w", err)
	}
	return boards, nil
}

func (r *Boards) Get(ctx context.Context, boardID pgtype.UUID) (BoardWithStats, error) {
	query := boardWithStatsBaseSelect + " WHERE b.id = $1"

	rows, err := r.db.Query(ctx, query, boardID)
	if err != nil {
		return BoardWithStats{}, fmt.Errorf("query board: %w", err)
	}
	defer rows.Close()

	board, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[BoardWithStats])
	if err != nil {
		return BoardWithStats{}, pgmap.TranslatePgErr(err)
	}
	return board, nil
}

func (r *Boards) Create(ctx context.Context, in CreateBoardInput) (BoardWithStats, error) {
	board, err := r.queries.CreateBoard(ctx, generated.CreateBoardParams{
		Title:       in.Title,
		Description: pgmap.PgTextFromString(in.Description),
		Size:        in.Size,
		AuthorID:    in.AuthorID,
	})
	if err != nil {
		return BoardWithStats{}, pgmap.TranslatePgErr(err)
	}
	// A freshly-created board has no votes and no games yet, so all stats
	// are zero. Skip a re-query.
	return BoardWithStats{Board: board}, nil
}

func (r *Boards) Delete(ctx context.Context, boardID pgtype.UUID) (generated.Board, error) {
	board, err := r.queries.DeleteBoard(ctx, boardID)
	if err != nil {
		return generated.Board{}, pgmap.TranslatePgErr(err)
	}
	return board, nil
}

func (r *Boards) TotalGamesPlayed(ctx context.Context, boardID pgtype.UUID) (generated.GetTotalGamesPlayedForBoardRow, error) {
	playedGames, err := r.queries.GetTotalGamesPlayedForBoard(ctx, boardID)
	if err != nil {
		return generated.GetTotalGamesPlayedForBoardRow{}, pgmap.TranslatePgErr(err)
	}
	return playedGames, nil
}
