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

// ===== Handler =====

type UsersHandler struct {
	users *service.Users
	votes *service.Votes
}

func NewUsersHandler(users *service.Users, votes *service.Votes) *UsersHandler {
	return &UsersHandler{users: users, votes: votes}
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

	huma.Register(api, huma.Operation{
		OperationID: "update-current-user",
		Method:      http.MethodPatch,
		Path:        "/users/me",
		Summary:     "Update the current User's username and/or email",
		Description: "Convenience alias for PATCH /users/{user_id} that " +
			"resolves the user id from the session cookie. The `email` " +
			"field is tri-state: omit the key to leave the column " +
			"unchanged, send `null` to clear it, or send a new address to " +
			"set it (which also resets verification and triggers a " +
			"verification mail).",
		Tags: []string{"Users"},
	}, h.updateMe)

	huma.Register(api, huma.Operation{
		OperationID: "list-votes-for-a-user",
		Method:      http.MethodGet,
		Path:        "/users/{user_id}/votes",
		Summary:     "List votes for a User",
		Tags:        []string{"Users"},
	}, h.listVotesFromUser)

	huma.Register(api, huma.Operation{
		OperationID: "list-votes-for-current-user",
		Method:      http.MethodGet,
		Path:        "/users/me/votes",
		Summary:     "List votes for current User",
		Tags:        []string{"Users"},
	}, h.listVotesFromMe)
}

func (h *UsersHandler) me(ctx context.Context, _ *struct{}) (*userOutput, error) {
	user, err := h.users.Me(ctx)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get me")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *UsersHandler) userByID(ctx context.Context, in *userByIDInput) (*userOutput, error) {
	user, err := h.users.UserByID(ctx, in.UserID)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get user by ID")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *UsersHandler) update(ctx context.Context, in *updateUserInput) (*userOutput, error) {
	user, err := h.users.UpdateUser(ctx, in.UserID, service.UpdateUserInput{
		Username: in.Body.Username,
		EmailSet: in.Body.Email.Set,
		Email:    in.Body.Email.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to update user")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *UsersHandler) updateMe(ctx context.Context, in *updateMeInput) (*userOutput, error) {
	user, err := h.users.UpdateMe(ctx, service.UpdateUserInput{
		Username: in.Body.Username,
		EmailSet: in.Body.Email.Set,
		Email:    in.Body.Email.Value,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to update current user")
	}

	return &userOutput{Body: userToOutput(user)}, nil
}

func (h *UsersHandler) listVotesFromUser(ctx context.Context, in *listVotesFromUserInput) (*listVotesFromUserOutput, error) {
	votes, err := h.votes.ListVotesFromUser(ctx, in.UserID.Value)
	if err != nil {
		return nil, toHumaErr(err, "", "failed to list votes for user")
	}

	body := make([]listVotesFromUserOutputBody, len(votes))
	for i, vote := range votes {
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
	for i, vote := range votes {
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
