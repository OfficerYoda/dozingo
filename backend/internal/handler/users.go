package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/officeryoda/dozingo/internal/avatar"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

// ===== Input/Output types =====

type userByIDInput struct {
	UserID string `path:"user_id"  format:"uuid"`
}

type deleteUserInputBody struct {
	Password string `json:"password" required:"true" minLength:"8" maxLength:"72"`
}

type deleteUserInput struct {
	Body deleteUserInputBody
}

// updateUserInputBody is a tri-state PATCH payload:
//
//   - username: omit the key entirely to leave the username unchanged.
//     Provide a non-empty string to rename. JSON null is rejected by the
//     schema (the field is non-nullable).
//   - email: omit the key (or send JSON null) to leave the email unchanged.
//     Send "" (empty string) to clear it. Send a valid address to set a new
//     one; the server will reset email_verified_at and dispatch a verification
//     mail to the new address.
type updateUserInputBody struct {
	Username *string `json:"username,omitempty" maxLength:"200"`
	Email    *string `json:"email,omitempty" maxLength:"200"`
}

type updateMeInput struct {
	Body updateUserInputBody
}

type listVotesFromUserInput struct {
	UserID types.UUIDParam `path:"user_id" format:"uuid"`
}

type listVotesFromUserOutputBody struct {
	VoteID        string `json:"vote_id"`
	VoteValue     int32  `json:"vote_value"`
	BoardID       string `json:"board_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Size          int32  `json:"size"`
	BoardAuthorID string `json:"board_author_id"`
	Score         int64  `json:"vote_score"`
	VoteCount     int64  `json:"vote_count"`
	PlayCount     int64  `json:"play_count"`
}

type listVotesFromUserOutput struct {
	Body []listVotesFromUserOutputBody
}

type avatarUploadInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Avatar huma.FormFile `form:"avatar" required:"true" doc:"The avatar image file. Allowed types: PNG, JPEG, WEBP. Max size: 20MB."`
	}]
}

// ===== Handler =====

type UsersHandler struct {
	users      *service.Users
	votes      *service.Votes
	avatarURLs *avatar.URLBuilder
}

func NewUsersHandler(users *service.Users, votes *service.Votes, avatarURLs *avatar.URLBuilder) *UsersHandler {
	return &UsersHandler{users: users, votes: votes, avatarURLs: avatarURLs}
}

func (h *UsersHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user",
		Method:      http.MethodGet,
		Path:        "/users/me",
		Summary:     "Information about current user",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.me)

	huma.Register(api, huma.Operation{
		OperationID: "delete-current-user",
		Method:      http.MethodDelete,
		Path:        "/users/me",
		Summary:     "Delete the current user",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.delete)

	huma.Register(api, huma.Operation{
		OperationID: "get-user-by-id",
		Method:      http.MethodGet,
		Path:        "/users/{user_id}",
		Summary:     "Get User by ID",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.userByID)

	huma.Register(api, huma.Operation{
		OperationID: "update-current-user",
		Method:      http.MethodPatch,
		Path:        "/users/me",
		Summary:     "Update the current User's username and/or email",
		Description: "Resolves the user id from the session cookie. The " +
			"`email` field is tri-state: omit the key (or send `null`) to " +
			"leave the column unchanged, send `\"\"` (empty string) to clear " +
			"it, or send a valid address to set it (which also resets " +
			"verification and triggers a verification mail).",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.StrictAuthLimiter)},
	}, h.updateMe)

	huma.Register(api, huma.Operation{
		OperationID: "list-votes-for-a-user",
		Method:      http.MethodGet,
		Path:        "/users/{user_id}/votes",
		Summary:     "List votes for a User",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadListLimiter)},
	}, h.listVotesFromUser)

	huma.Register(api, huma.Operation{
		OperationID: "list-votes-for-current-user",
		Method:      http.MethodGet,
		Path:        "/users/me/votes",
		Summary:     "List votes for current User",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadListLimiter)},
	}, h.listVotesFromMe)

	huma.Register(api, huma.Operation{
		OperationID: "update-current-user-avatar",
		Method:      http.MethodPut,
		Path:        "/users/me/avatar",
		Summary:     "Upload a new profile picture",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteHeavyLimiter)},
	}, h.uploadAvatar)
}

func (h *UsersHandler) me(ctx context.Context, _ *struct{}) (*userOutput, error) {
	user, err := h.users.Me(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get me")
	}

	return &userOutput{Body: userToOutput(user, h.avatarURLs)}, nil
}

func (h *UsersHandler) delete(ctx context.Context, in *deleteUserInput) (*struct{}, error) {
	err := h.users.Delete(ctx, in.Body.Password)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to delete user")
	}

	return nil, nil
}

func (h *UsersHandler) userByID(ctx context.Context, in *userByIDInput) (*userOutput, error) {
	user, err := h.users.UserByID(ctx, in.UserID)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get user by ID")
	}

	return &userOutput{Body: userToOutput(user, h.avatarURLs)}, nil
}

func (h *UsersHandler) updateMe(ctx context.Context, in *updateMeInput) (*userOutput, error) {
	user, err := h.users.UpdateMe(ctx, service.UpdateUserInput{
		Username: in.Body.Username,
		Email:    in.Body.Email,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to update current user")
	}

	return &userOutput{Body: userToOutput(user, h.avatarURLs)}, nil
}

func (h *UsersHandler) listVotesFromUser(ctx context.Context, in *listVotesFromUserInput) (*listVotesFromUserOutput, error) {
	votes, err := h.votes.ListVotesFromUser(ctx, in.UserID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list votes for user")
	}

	body := make([]listVotesFromUserOutputBody, len(votes))
	for i := range votes {
		vote := &votes[i]
		body[i] = listVotesFromUserOutputBody{
			VoteID:        vote.VoteID.String(),
			VoteValue:     vote.VoteValue,
			BoardID:       vote.BoardID.String(),
			Title:         vote.Title,
			Description:   vote.Description.String,
			Size:          vote.Size,
			BoardAuthorID: vote.BoardAuthorID.String(),
			Score:         vote.Score,
			VoteCount:     vote.VoteCount,
			PlayCount:     vote.PlayCount,
		}
	}

	return &listVotesFromUserOutput{Body: body}, nil
}

func (h *UsersHandler) listVotesFromMe(ctx context.Context, _ *struct{}) (*listVotesFromUserOutput, error) {
	votes, err := h.votes.ListVotesFromMe(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list votes for current user")
	}

	body := make([]listVotesFromUserOutputBody, len(votes))
	for i := range votes {
		vote := &votes[i]
		body[i] = listVotesFromUserOutputBody{
			VoteID:        vote.VoteID.String(),
			VoteValue:     vote.VoteValue,
			BoardID:       vote.BoardID.String(),
			Title:         vote.Title,
			Description:   vote.Description.String,
			Size:          vote.Size,
			BoardAuthorID: vote.BoardAuthorID.String(),
			Score:         vote.Score,
			VoteCount:     vote.VoteCount,
			PlayCount:     vote.PlayCount,
		}
	}

	return &listVotesFromUserOutput{Body: body}, nil
}

func (h *UsersHandler) uploadAvatar(ctx context.Context, in *avatarUploadInput) (*userOutput, error) {
	user, err := h.users.UploadAvatar(ctx, in.RawBody.Data().Avatar)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to upload avatar")
	}

	return &userOutput{Body: userToOutput(user, h.avatarURLs)}, nil
}
