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
	state, err := mgr.StateProvider.Refresh(params.Name)
	if err != nil {
		return UserResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}
	user, proverr := providers.GetUser(&state, api.User{Name: params.Name})
	if proverr == providers.ProviderOk {
		return UserResponse{Success: true, Created: false, User: user}, nil
	}

	if proverr == providers.ProviderNotFound {
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

	state, err := mgr.StateProvider.Refresh(params.Name)
	if err != nil {
		return UserResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	new := api.NewDefaultUser(params.Name)
	proverr := providers.CreateUser(&state, new)
	if proverr != providers.ProviderOk {
		return UserResponse{}, mgr.MakeNewResponseError(ProviderError, proverr, "Unable to create user: %s", params.Name)
	}
	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return UserResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to save user: %s", params.Name)
	}

	return UserResponse{Success: true, Created: true, User: new}, nil
}
