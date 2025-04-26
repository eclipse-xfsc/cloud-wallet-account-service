package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/env"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
)

var logger = common.GetLogger()

func getTokenValue(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	res, _ := strings.CutPrefix(token, "Bearer ")
	return res
}

func CheckExistenceAndGetUserData(dataFetcher ...services.DataFetcher) gin.HandlerFunc {
	var oidcProvider *services.OIDCProvider
	if len(dataFetcher) > 0 {
		oidcProvider = services.GetOidcProvider(dataFetcher[0])
	} else {
		oidcProvider = services.GetOidcProvider()
	}
	return func(c *gin.Context) {
		if common.CheckSkipEndpointAuth(c) {
			c.Next()
			return
		}
		token := getTokenValue(c)
		user, err := oidcProvider.GetUser(token)
		if user != nil {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), common.UserKey, user))
			c.Next()
		} else {
			if err != nil {
				_ = c.Error(err)
			}
			c.JSON(http.StatusForbidden, gin.H{"message": common.UserNotFound})
			c.Abort()
		}
	}
}

func CreateCryptoKeysIfAccountIsNew(envir ...common.Env) gin.HandlerFunc {
	var e common.Env
	if len(envir) == 0 {
		e = env.GetEnv()
	} else {
		e = envir[0]
	}
	return func(c *gin.Context) {
		if common.CheckSkipEndpointAuth(c) {
			c.Next()
			return
		}
		var usInf = c.Request.Context().Value(common.UserKey)
		if usInf == nil {
			c.JSON(http.StatusForbidden, gin.H{"message": common.UserNotFound})
			c.Abort()
			return
		}
		userInfo := usInf.(*common.UserInfo)
		if secret := model.GetUserSecretIdDbEntry(e.GetDB(), userInfo.ID()); secret == nil || (secret != nil && secret.SecretId == "" && secret.UserId == "") {
			secretId := model.CreateUserSecretId(userInfo.ID())
			cryptoProvider := e.GetCryptoProvider()
			generateSecret(c, secretId, cryptoProvider, e.GetNamespace())
			model.CreateUserSecretDbEntry(e.GetDB(), userInfo.ID(), secretId)
		}
		c.Next()
	}
}

func generateSecret(c *gin.Context, secretId string, cryptoProvider types.CryptoProvider, namespace string) {
	ctx := context.Background()
	identifier := types.CryptoIdentifier{
		KeyId: secretId,
		CryptoContext: types.CryptoContext{
			Namespace: namespace,
			Context:   ctx,
		},
	}
	err := cryptoProvider.GenerateKey(types.CryptoKeyParameter{Identifier: identifier, KeyType: types.Ed25519})
	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{"message": common.CryptoError})
		c.Abort()
		return
	}
}
