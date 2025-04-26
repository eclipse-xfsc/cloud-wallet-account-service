package test

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/mock"
)

type ProviderMock struct {
	mock.Mock
}

func (k *ProviderMock) GetUserInfo(ctx context.Context, token string, realm string) (*gocloak.UserInfo, error) {
	args := k.Called(ctx, token, realm)
	us := args.Get(0)
	if us == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.UserInfo), nil
}
