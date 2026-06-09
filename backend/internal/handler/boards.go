package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

// ===== Input/Output types =====

type boardOutput struct {
	BoardID     string  `json:"board_id" format:"uuid"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Size        int32   `json:"size"`
	AuthorID    string  `json:"author_id" format:"uuid"`
	Score       int64   `json:"score"`
	VoteCount   int64   `json:"vote_count"`
	PlayCount   int64   `json:"play_count"`
}

type getBoardsInput struct {
	AuthorID string `query:"author_id"`
	Size     int32  `query:"size"`
	Sort     string `query:"sort" enum:"newest,oldest,most-liked,least-liked,most-played,least-played" default:"newest"`
	Limit    int32  `query:"limit" minimum:"1" default:"20"`
	Search   string `query:"search"`
}

type getBoardsBySessionInput struct {
	Size  int32  `query:"size"`
	Sort  string `query:"sort" enum:"newest,oldest,most-liked,least-liked,most-played,least-played" default:"newest"`
	Limit int32  `query:"limit" minimum:"1" default:"20"`
}

type getBoardsOutput struct {
	Body []boardOutput
}

type getBoardByIDInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type getBoardByIDOutput struct {
	Body boardOutput
}

type createBoardInputBody struct {
	Title       string  `json:"title" required:"true" maxLength:"200"`
	Description *string `json:"description,omitempty" maxLength:"500"`
	Size        int32   `json:"size" required:"true" minimum:"4" maximum:"6"`
}

type createBoardInput struct {
	Body createBoardInputBody
}

type createBoardOutput struct {
	Body boardOutput
}

type deleteBoardInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type getTotalGamesPlayedInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type getTotalGamesPlayedBody struct {
	BoardID    string `json:"board_id" format:"uuid"`
	BoardTitle string `json:"board_title"`
	TotalGames int64  `json:"total_games"`
}

type getTotalGamesPlayedOutput struct {
	Body getTotalGamesPlayedBody
}

// ===== Handler =====

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
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadListLimiter)},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-boards-from-session",
		Method:      http.MethodGet,
		Path:        "/me/boards",
		Summary:     "Get all boards of current user",
		Tags:        []string{"Boards"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadListLimiter)},
	}, h.listBySession)

	huma.Register(api, huma.Operation{
		OperationID: "get-board-by-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}",
		Summary:     "Get a board by ID",
		Tags:        []string{"Boards"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "create-board",
		Method:      http.MethodPost,
		Path:        "/boards",
		Summary:     "Create a board",
		Tags:        []string{"Boards"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "delete-board",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}",
		Summary:     "Delete a board",
		Tags:        []string{"Boards"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.delete)

	huma.Register(api, huma.Operation{
		OperationID: "total-played-games",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/total-played-games",
		Summary:     "Get total count of played games",
		Tags:        []string{"Boards"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.totalPlayedGames)
}

func (h *BoardsHandler) list(ctx context.Context, in *getBoardsInput) (*getBoardsOutput, error) {
	boards, err := h.svc.List(ctx, service.BoardListFilter{
		AuthorID: in.AuthorID,
		Size:     in.Size,
		Sort:     in.Sort,
		Limit:    in.Limit,
		Search:   in.Search,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list boards")
	}

	return &getBoardsOutput{Body: mapSlice(boards, boardToOutput)}, nil
}

func (h *BoardsHandler) listBySession(ctx context.Context, in *getBoardsBySessionInput) (*getBoardsOutput, error) {
	boards, err := h.svc.ListBySession(ctx, service.BoardListFilter{
		Size:  in.Size,
		Sort:  in.Sort,
		Limit: in.Limit,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list boards")
	}

	return &getBoardsOutput{Body: mapSlice(boards, boardToOutput)}, nil
}

func (h *BoardsHandler) get(ctx context.Context, in *getBoardByIDInput) (*getBoardByIDOutput, error) {
	board, err := h.svc.Get(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "board not found", "failed to get board")
	}

	return &getBoardByIDOutput{Body: boardToOutput(board)}, nil
}

func (h *BoardsHandler) create(ctx context.Context, in *createBoardInput) (*createBoardOutput, error) {
	board, err := h.svc.Create(ctx, service.CreateBoardInput{
		Title:       in.Body.Title,
		Description: in.Body.Description,
		Size:        in.Body.Size,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to create board")
	}

	return &createBoardOutput{Body: boardToOutput(board)}, nil
}

func (h *BoardsHandler) delete(ctx context.Context, in *deleteBoardInput) (*struct{}, error) {
	err := h.svc.Delete(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "board not found", "failed to delete board")
	}

	return &struct{}{}, nil
}

func (h *BoardsHandler) totalPlayedGames(ctx context.Context, in *getTotalGamesPlayedInput) (*getTotalGamesPlayedOutput, error) {
	playedGames, err := h.svc.TotalGamesPlayed(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to count total played games")
	}

	return &getTotalGamesPlayedOutput{
		Body: getTotalGamesPlayedBody{
			BoardID:    playedGames.BoardID.String(),
			BoardTitle: playedGames.BoardTitle,
			TotalGames: playedGames.TotalGames,
		},
	}, nil
}

func boardToOutput(board repository.BoardWithStats) boardOutput {
	return boardOutput{
		BoardID:     board.ID.String(),
		Title:       board.Title,
		Description: pgmap.StringFromPgText(board.Description),
		Size:        board.Size,
		AuthorID:    board.AuthorID.String(),
		Score:       board.Score,
		VoteCount:   board.VoteCount,
		PlayCount:   board.PlayCount,
	}
}
