package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func PluginRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {
	pluginGroup := group.Group("/plugin-discovery")

	pluginGroup.GET("", common.ConstructResponse(handlers.ListPlugins, e))

	return pluginGroup
}
