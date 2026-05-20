package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
)

// Boards is the persistence repository for the boards aggregate.
type Boards struct {
	q  *generated.Queries
	db generated.DBTX
}

// BoardListFilter narrows the result of List. Zero-valued fields are ignored.
//
// AuthorID accepts the raw query-string form (a UUID or empty string) so the
// repository, not the handler, owns parsing. An invalid UUID results in
// domain.ErrInvalid.
type BoardListFilter struct {
	AuthorID string
	Size     int32
}

// CreateBoardInput is the repository-level input for creating a board.
//
// Description is *string so the caller can express "no value provided" without
// leaking pgtype into the service layer.
type CreateBoardInput struct {
	Title       string
	Description *string
	Size        int32
	AuthorID    pgtype.UUID
}

// List returns boards matching the given filters, ordered by created_at desc.
//
// This used to be a hand-rolled SQL builder living inside the HTTP handler.
// It now lives at the persistence seam where it belongs; an even cleaner
// follow-up is to express the filtered query in db/queries/boards.sql so
// sqlc generates it for us.
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
			return nil, fmt.Errorf("invalid author_id: %w", domain.ErrInvalid)
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

// Get returns the board with the given id, or domain.ErrNotFound.
func (r *Boards) Get(ctx context.Context, id pgtype.UUID) (generated.Board, error) {
	board, err := r.q.GetBoardByID(ctx, id)
	if err != nil {
		return generated.Board{}, translatePgErr(err)
	}
	return board, nil
}

// Create inserts a board and returns the persisted row. A unique-constraint
// violation is reported as domain.ErrConflict.
func (r *Boards) Create(ctx context.Context, in CreateBoardInput) (generated.Board, error) {
	board, err := r.q.CreateBoard(ctx, generated.CreateBoardParams{
		Title:       in.Title,
		Description: pgTextFromString(in.Description),
		Size:        in.Size,
		AuthorID:    in.AuthorID,
	})
	if err != nil {
		return generated.Board{}, translatePgErr(err)
	}
	return board, nil
}

// Delete removes the board with the given id and returns the deleted row.
// Returns domain.ErrNotFound if no row matched.
func (r *Boards) Delete(ctx context.Context, id pgtype.UUID) (generated.Board, error) {
	board, err := r.q.DeleteBoard(ctx, id)
	if err != nil {
		return generated.Board{}, translatePgErr(err)
	}
	return board, nil
}

// translatePgErr maps pgx/pgconn errors to domain sentinels. Anything we don't
// recognize is returned unchanged so it surfaces as a 500 in the handler.
func translatePgErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w", domain.ErrNotFound)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrConflict)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, domain.ErrInvalid)
		}
	}
	return err
}
