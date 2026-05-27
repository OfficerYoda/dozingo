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

// sortMode classifies the requested sort into the kind of FROM clause and
// ORDER BY expression it needs.
type sortMode struct {
	join    string // optional LEFT JOIN ... clause
	groupBy bool   // whether GROUP BY b.id is required
	orderBy string // ORDER BY expression (without "ORDER BY" prefix)
}

func resolveSort(sort string) sortMode {
	switch sort {
	case "oldest":
		return sortMode{orderBy: "b.created_at ASC"}
	case "most-liked":
		return sortMode{
			join:    "LEFT JOIN votes v ON v.board_id = b.id",
			groupBy: true,
			orderBy: "COALESCE(SUM(v.vote_value), 0) DESC, b.created_at DESC",
		}
	case "least-liked":
		return sortMode{
			join:    "LEFT JOIN votes v ON v.board_id = b.id",
			groupBy: true,
			orderBy: "COALESCE(SUM(v.vote_value), 0) ASC, b.created_at DESC",
		}
	case "most-played":
		return sortMode{
			join:    "LEFT JOIN games g ON g.board_id = b.id",
			groupBy: true,
			orderBy: "COUNT(g.id) DESC, b.created_at DESC",
		}
	case "least-played":
		return sortMode{
			join:    "LEFT JOIN games g ON g.board_id = b.id",
			groupBy: true,
			orderBy: "COUNT(g.id) ASC, b.created_at DESC",
		}
	default:
		// "newest" and any unknown value fall back to newest-first.
		return sortMode{orderBy: "b.created_at DESC"}
	}
}

func (r *Boards) List(ctx context.Context, f BoardListFilter) ([]generated.Board, error) {
	var (
		query strings.Builder
		args  []any
		i     = 1
	)

	mode := resolveSort(f.Sort)

	query.WriteString("SELECT b.id, b.title, b.size, b.author_id, b.created_at, b.updated_at, b.description FROM boards b")
	if mode.join != "" {
		query.WriteString(" ")
		query.WriteString(mode.join)
	}
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

	if mode.groupBy {
		query.WriteString(" GROUP BY b.id")
	}

	query.WriteString(" ORDER BY ")
	query.WriteString(mode.orderBy)

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

	boards, err := pgx.CollectRows(rows, pgx.RowToStructByName[generated.Board])
	if err != nil {
		return nil, fmt.Errorf("scan boards: %w", err)
	}
	return boards, nil
}

func (r *Boards) Get(ctx context.Context, boardID pgtype.UUID) (generated.Board, error) {
	board, err := r.queries.GetBoardByID(ctx, boardID)
	if err != nil {
		return generated.Board{}, pgmap.TranslatePgErr(err)
	}
	return board, nil
}

func (r *Boards) Create(ctx context.Context, in CreateBoardInput) (generated.Board, error) {
	board, err := r.queries.CreateBoard(ctx, generated.CreateBoardParams{
		Title:       in.Title,
		Description: pgmap.PgTextFromString(in.Description),
		Size:        in.Size,
		AuthorID:    in.AuthorID,
	})
	if err != nil {
		return generated.Board{}, pgmap.TranslatePgErr(err)
	}
	return board, nil
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
