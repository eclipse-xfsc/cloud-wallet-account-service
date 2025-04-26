package services

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
	"github.com/ReneKroon/ttlcache"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
)

var logger = common.GetLogger()

var users = newCache()

func newCache() *ttlcache.Cache {
	cache := ttlcache.NewCache()
	cache.SetTTL(config.ServerConfiguration.KeyCloak.TokenTTL)
	return cache
}

type OIDCProvider struct {
	*gocloak.GoCloak
	fetcher DataFetcher
}

type DataFetcher interface {
	GetUserInfo(ctx context.Context, accessToken string, realm string) (*gocloak.UserInfo, error)
}

func (p *OIDCProvider) GetUser(token string) (*common.UserInfo, error) {
	if user, ok := users.Get(token); ok {
		logger.Debug("hit cache")
		return user.(*common.UserInfo), nil
	}
	user, err := p.fetcher.GetUserInfo(context.Background(), token, config.ServerConfiguration.KeyCloak.RealmName)
	logger.Debug("fetched user data", "userIsFound", user != nil, "error", err)
	if user != nil {
		userInfo := common.UserInfo{user}
		users.Set(token, &userInfo)
		return &userInfo, nil
	}
	return nil, err
}

var oidcProvider *OIDCProvider

func GetOidcProvider(f ...DataFetcher) *OIDCProvider {
	if oidcProvider != nil {
		return oidcProvider
	}

	oidcService := gocloak.NewClient(config.ServerConfiguration.KeyCloak.Url, func(cloak *gocloak.GoCloak) {
		// additional hooks
	})

	if len(f) == 0 {
		f = []DataFetcher{oidcService}
	}
	oidcProvider = &OIDCProvider{oidcService, f[0]}
	return oidcProvider
}
