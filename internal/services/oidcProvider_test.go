package services

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

var testToken = "testtokentesttoken"
var testSub = "testID"

type ProviderMock struct {
	mock.Mock
}

func (k *ProviderMock) GetUserInfo(ctx context.Context, token string, realm string) (*gocloak.UserInfo, error) {
	if token == testToken {
		return &gocloak.UserInfo{Sub: &testSub}, nil
	}
	return nil, fmt.Errorf("user not found")
}

func TestOIDCProvider_GetUser(t *testing.T) {
	mockObj := &ProviderMock{}
	provider := GetOidcProvider(mockObj)
	user, err := provider.GetUser(testToken)
	require.Nil(t, err)
	require.Equal(t, testSub, user.ID())
}
