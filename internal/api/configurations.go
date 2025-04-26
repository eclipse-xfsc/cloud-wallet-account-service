package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func ConfiguationsRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {
	configurationsGroup := group.Group("/configurations")

	configurationsGroup.GET("/list", common.ConstructResponse(handlers.GetConfigurations, e))
	configurationsGroup.GET("/getUserInfo", common.ConstructResponse(handlers.GetUserInfo, e))
	configurationsGroup.PUT("/save", common.ConstructResponse(handlers.SaveConfigurations, e))

	return configurationsGroup
}
