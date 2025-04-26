package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func KmsRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {

	kmsGroup := group.Group("/kms")

	kmsGroup.GET("/keyTypes", common.ConstructResponse(handlers.GetSupportedKeysAlgs, e))

	didGroup := kmsGroup.Group("/did")

	didGroup.GET("/list", common.ConstructResponse(handlers.ListDID, e))

	didGroup.POST("/create", common.ConstructResponse(handlers.CreateDID, e))

	didGroup.DELETE("/:kid", common.ConstructResponse(handlers.DeleteDID, e))

	return kmsGroup
}
