package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

/// ===== Input/Output types =====

type BoardOutput struct {
	BoardID     string  `json:"board_id" format:"uuid"`
	Title       string  `json:"title" format:"text"`
	Description *string `json:"description" format:"text"`
	Size        int32   `json:"size" format:"integer"`
	AuthorID    string  `json:"author_id" format:"uuid"`
}

type GetBoardsInput struct {
	AuthorID string `query:"author_id"`
	Size     int32  `query:"size"`
}

type GetBoardsOutput struct {
	Body []BoardOutput
}

type GetBoardByIDInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type GetBoardByIDOutput struct {
	Body BoardOutput
}

type CreateBoardInput struct {
	Body struct {
		Title       string          `json:"title" format:"text" required:"true" maxLength:"200"`
		Description *string         `json:"description,omitempty" format:"text" maxLength:"500"`
		Size        int32           `json:"size" format:"integer" required:"true" maxLength:"200"`
		AuthorID    types.UUIDParam `json:"author_id" format:"uuid" required:"true"`
	}
}

type CreateBoardOutput struct {
	Body BoardOutput
}

type DeleteBoardInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

/// ===== Handler =====

type BoardsHandler struct {
	svc *service.Boards
}

func NewBoardsHandler(svc *service.Boards) *BoardsHandler {
	return &BoardsHandler{svc: svc}
}

func (h *BoardsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-boards",
		Method:      http.MethodGet,
		Path:        "/boards",
		Summary:     "Get all boards with optional filters",
		Tags:        []string{"Boards"},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-board-by-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}",
		Summary:     "Get a board by ID",
		Tags:        []string{"Boards"},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "create-board",
		Method:      http.MethodPost,
		Path:        "/boards",
		Summary:     "Create a board",
		Tags:        []string{"Boards"},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "delete-board",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}",
		Summary:     "Delete a board",
		Tags:        []string{"Boards"},
	}, h.delete)
}

func (h *BoardsHandler) list(ctx context.Context, in *GetBoardsInput) (*GetBoardsOutput, error) {
	boards, err := h.svc.List(ctx, repository.BoardListFilter{
		AuthorID: in.AuthorID,
		Size:     in.Size,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list boards")
	}
	return &GetBoardsOutput{Body: mapSlice(boards, boardToOutput)}, nil
}

func (h *BoardsHandler) get(ctx context.Context, in *GetBoardByIDInput) (*GetBoardByIDOutput, error) {
	board, err := h.svc.Get(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "board not found", "failed to get board")
	}
	return &GetBoardByIDOutput{Body: boardToOutput(board)}, nil
}

func (h *BoardsHandler) create(ctx context.Context, in *CreateBoardInput) (*CreateBoardOutput, error) {
	board, err := h.svc.Create(ctx, repository.CreateBoardInput{
		Title:       in.Body.Title,
		Description: in.Body.Description,
		Size:        in.Body.Size,
		AuthorID:    in.Body.AuthorID.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create board")
	}
	return &CreateBoardOutput{Body: boardToOutput(board)}, nil
}

func (h *BoardsHandler) delete(ctx context.Context, in *DeleteBoardInput) (*struct{}, error) {
	if err := h.svc.Delete(ctx, in.BoardID.Value); err != nil {
		return nil, toHumaErr(err, "board not found", "failed to delete board")
	}
	return &struct{}{}, nil
}

func boardToOutput(board generated.Board) BoardOutput {
	return BoardOutput{
		BoardID:     board.ID.String(),
		Title:       board.Title,
		Description: pgmap.StringFromPgText(board.Description),
		Size:        board.Size,
		AuthorID:    board.AuthorID.String(),
	}
}
