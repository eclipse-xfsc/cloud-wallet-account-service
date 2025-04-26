package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func CredentialRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {

	credentialGroup := group.Group("/credentials")

	credentialGroup.GET("/list", common.ConstructResponse(handlers.ListCredentials, e))
	credentialGroup.POST("/list", common.ConstructResponse(handlers.ListCredentials, e))

	credentialGroup.GET("/history", common.ConstructResponse(handlers.ListCredentials, e))

	credentialGroup.DELETE("/:id", common.ConstructResponse(handlers.DeleteCredential, e))

	credentialGroup.GET("/:id/revoke", common.ConstructResponse(handlers.RevokeCredential, e))

	credentialGroup.GET("/schemas", common.ConstructResponse(handlers.GetCredentialConfigurations, e))
	credentialGroup.POST("/issue", common.ConstructResponse(handlers.RequestIssuance, e))

	CredentialBackupRoutes(credentialGroup, e)
	CredentialOfferRoutes(credentialGroup, e)

	return credentialGroup
}
