package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func SettingsRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {

	settingsGroup := group.Group("/settings")

	settingsGroup.GET("/ui", common.ConstructResponse(handlers.GetUiSettings, e))

	settingsGroup.POST("/ui", common.ConstructResponse(handlers.SetUiSettings, e))

	return settingsGroup
}
