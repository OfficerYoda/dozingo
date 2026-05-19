package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Input/Output types =====

type BoardOutput struct {
	ID          string `json:"id" format:"uuid"`
	Title       string `json:"title" format:"text"`
	Description string `json:"description" format:"text"`
	Size        int32  `json:"size" format:"integer"`
	AuthorID    string `json:"author_id" format:"uuid"`
}

type GetBoardsInput struct {
	AuthorID string `query:"author_id"`
	Size     int32  `query:"size"`
}

type GetBoardsOutput struct {
	Body []BoardOutput
}

type GetBoardByIDInput struct {
	ID string `path:"board_id" format:"uuid"`
}

type GetBoardByIDOutput struct {
	Body BoardOutput
}

type CreateBoardInput struct {
	Body struct {
		Title       string  `json:"title" format:"text" required:"true" maxLength:"200"`
		Description *string `json:"description,omitempty" format:"text" maxLength:"500"`
		Size        int32   `json:"size" format:"integer" required:"true" maxLength:"200"`
		AuthorID    string  `json:"author_id" format:"uuid" required:"true"`
	}
}

type CreateBoardOutput struct {
	Body BoardOutput
}

type DeleteBoardInput struct {
	ID string `path:"board_id" format:"uuid"`
}

/// ===== Register =====

func RegisterBoards(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-boards",
		Method:      http.MethodGet,
		Path:        "/boards",
		Summary:     "Get all boards with optional filters",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *GetBoardsInput) (*GetBoardsOutput, error) {
		return getBoards(ctx, pool, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-board-by-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}",
		Summary:     "Get a board by ID",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *GetBoardByIDInput) (*GetBoardByIDOutput, error) {
		return getBoardByID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-board",
		Method:      http.MethodPost,
		Path:        "/boards",
		Summary:     "Create a board",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *CreateBoardInput) (*CreateBoardOutput, error) {
		return createBoard(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-board",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}",
		Summary:     "Delete a board",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *DeleteBoardInput) (*struct{}, error) {
		return deleteBoard(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getBoards(ctx context.Context, pool *pgxpool.Pool, input GetBoardsInput) (*GetBoardsOutput, error) {
	rows, err := queryBoardsFiltered(ctx, input, pool)
	if err != nil {
		return nil, internalError(err, "failed to query boards")
	}
	defer rows.Close()

	boards, err := pgx.CollectRows(rows, pgx.RowToStructByName[generated.Board])
	if err != nil {
		return nil, internalError(err, "failed to scan boards")
	}

	return &GetBoardsOutput{Body: mapSlice(boards, boardToOutput)}, nil
}

func getBoardByID(ctx context.Context, queries *generated.Queries, input GetBoardByIDInput) (*GetBoardByIDOutput, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	board, err := queries.GetBoardByID(ctx, id)
	if err != nil {
		return nil, notFoundOr500(err, "board not found", "failed to get board")
	}

	return &GetBoardByIDOutput{Body: boardToOutput(board)}, nil
}

func createBoard(ctx context.Context, queries *generated.Queries, input CreateBoardInput) (*CreateBoardOutput, error) {
	authorID, err := uuidFromString(input.Body.AuthorID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid author_id", err)
	}

	description := pgTextFromString(input.Body.Description)

	board, err := queries.CreateBoard(ctx, generated.CreateBoardParams{
		Title:       input.Body.Title,
		Description: description,
		Size:        input.Body.Size,
		AuthorID:    authorID,
	})
	if err != nil {
		return nil, internalError(err, "failed to create board")
	}

	return &CreateBoardOutput{Body: boardToOutput(board)}, nil
}

func deleteBoard(ctx context.Context, queries *generated.Queries, input DeleteBoardInput) (*struct{}, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	_, err = queries.DeleteBoard(ctx, id)
	if err != nil {
		return nil, notFoundOr500(err, "board not found", "failed to delete board")
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func queryBoardsFiltered(ctx context.Context, input GetBoardsInput, pool *pgxpool.Pool) (pgx.Rows, error) {
	query := "SELECT * FROM boards WHERE 1=1"
	args := []any{}
	i := 1

	if input.AuthorID != "" {
		query += fmt.Sprintf(" AND author_id = $%d", i)
		args = append(args, input.AuthorID)
		i++
	}

	if input.Size != 0 {
		query += fmt.Sprintf(" AND size = $%d", i)
		args = append(args, input.Size)
	}

	query += " ORDER BY created_at DESC;"

	rows, err := pool.Query(ctx, query, args...)

	return rows, err
}

func boardToOutput(board generated.Board) BoardOutput {
	return BoardOutput{
		ID:          board.ID.String(),
		Title:       board.Title,
		Description: board.Description.String,
		Size:        board.Size,
		AuthorID:    board.AuthorID.String(),
	}
}
