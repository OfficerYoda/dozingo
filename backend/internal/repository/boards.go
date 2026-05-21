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
}

type CreateBoardInput struct {
	Title       string
	Description *string
	Size        int32
	AuthorID    pgtype.UUID
}

func (r *Boards) List(ctx context.Context, f BoardListFilter) ([]generated.Board, error) {
	var (
		query strings.Builder
		args  []any
		i     = 1
	)
	query.WriteString("SELECT id, title, size, author_id, created_at, updated_at, description FROM boards WHERE 1=1")

	if f.AuthorID != "" {
		var authorUUID pgtype.UUID
		if err := authorUUID.Scan(f.AuthorID); err != nil {
			return nil, fmt.Errorf("invalid author_id: %w", domain.ErrBadInput)
		}
		fmt.Fprintf(&query, " AND author_id = $%d", i)
		args = append(args, authorUUID)
		i++
	}

	if f.Size != 0 {
		fmt.Fprintf(&query, " AND size = $%d", i)
		args = append(args, f.Size)
	}

	query.WriteString(" ORDER BY created_at DESC")

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

func (r *Boards) Get(ctx context.Context, id pgtype.UUID) (generated.Board, error) {
	board, err := r.queries.GetBoardByID(ctx, id)
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

func (r *Boards) Delete(ctx context.Context, id pgtype.UUID) (generated.Board, error) {
	board, err := r.queries.DeleteBoard(ctx, id)
	if err != nil {
		return generated.Board{}, pgmap.TranslatePgErr(err)
	}
	return board, nil
}
