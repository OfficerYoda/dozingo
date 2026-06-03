package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/service"
)

// ===== Input/Output types =====

type userByIDInput struct {
	UserID string `path:"user_id"  format:"uuid"`
}

// ===== Handler =====

type UsersHandler struct {
	svc *service.Users
}

func NewUsersHandler(svc *service.Users) *UsersHandler {
	return &UsersHandler{svc: svc}
}

func (h *UsersHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user",
		Method:      http.MethodGet,
		Path:        "/users/me",
		Summary:     "Information about current user",
		Tags:        []string{"Users"},
	}, h.me)

	huma.Register(api, huma.Operation{
		OperationID: "get-user-by-id",
		Method:      http.MethodGet,
		Path:        "/users/{user_id}",
		Summary:     "Get User by ID",
		Tags:        []string{"Users"},
	}, h.userByID)
}

func (h *UsersHandler) me(ctx context.Context, _ *struct{}) (*userOutput, error) {
	user, err := h.svc.Me(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get me")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *UsersHandler) userByID(ctx context.Context, in *userByIDInput) (*userOutput, error) {
	user, err := h.svc.UserByID(ctx, in.UserID)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get user by ID")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}
