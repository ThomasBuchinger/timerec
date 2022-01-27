package server

import (
	"context"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

type SearchUserParams struct {
	Name     string
	Inactive bool
}

type UserResponse struct {
	Success bool     `json:"success"`
	Created bool     `json:"created,omitempty"`
	User    api.User `json:"user"`
}

func (mgr *TimerecServer) GetUser(ctx context.Context, params SearchUserParams) (UserResponse, error) {
	user, err := mgr.StateProvider.GetUser(api.User{Name: params.Name})
	if err == nil {
		return UserResponse{Success: true, Created: false, User: user}, nil
	}

	if err.Error() == string(providers.ProviderErrorNotFound) {
		return UserResponse{Success: false, Created: false}, nil
	}

	return UserResponse{}, mgr.MakeNewResponseError(ServerError, err, "Error querying User '%s'", params.Name)
}

func (mgr *TimerecServer) CreateUserIfMissing(ctx context.Context, params SearchUserParams) (UserResponse, error) {
	resp, err := mgr.GetUser(ctx, params)
	if err != nil {
		mgr.Logger.Error(err)
		return UserResponse{}, err
	}
	if resp.Success {
		return resp, nil
	}

	new := api.NewDefaultUser(params.Name)
	saved, err := mgr.StateProvider.CreateUser(new)
	if err != nil {
		return UserResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to create user: %s", params.Name)
	}
	return UserResponse{Success: true, Created: true, User: saved}, nil
}
