package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

// ===== Input/Output types =====

type userByIDInput struct {
	UserID string `path:"user_id"  format:"uuid"`
}

// updateUserInputBody is a tri-state PATCH payload:
//
//   - username: omit the key entirely to leave the username unchanged.
//     Provide a non-empty string to rename. JSON null is rejected by the
//     schema (the field is non-nullable).
//   - email: omit the key entirely to leave the email unchanged. Send
//     explicit JSON null to clear it. Send a string to set a new address;
//     the server will reset email_verified_at and dispatch a verification
//     mail to the new address.
type updateUserInputBody struct {
	Username *string              `json:"username,omitempty" maxLength:"200"`
	Email    types.NullableString `json:"email,omitempty" format:"email" maxLength:"200"`
}

type updateUserInput struct {
	UserID string `path:"user_id" format:"uuid"`
	Body   updateUserInputBody
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

	huma.Register(api, huma.Operation{
		OperationID: "update-user",
		Method:      http.MethodPatch,
		Path:        "/users/{user_id}",
		Summary:     "Update a User's username and/or email",
		Description: "Partial update. Only the user themselves may edit their " +
			"own row. The `email` field is tri-state: omit the key to leave " +
			"the column unchanged, send `null` to clear it, or send a new " +
			"address to set it (which also resets verification and triggers " +
			"a verification mail).",
		Tags: []string{"Users"},
	}, h.update)
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

func (h *UsersHandler) update(ctx context.Context, in *updateUserInput) (*userOutput, error) {
	user, err := h.svc.UpdateUser(ctx, in.UserID, service.UpdateUserInput{
		Username: in.Body.Username,
		EmailSet: in.Body.Email.Set,
		Email:    in.Body.Email.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to update user")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}
