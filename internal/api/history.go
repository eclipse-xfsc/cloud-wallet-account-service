package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func HistoryRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {
	historyGroup := group.Group("/history")

	historyGroup.GET("/list", common.ConstructResponse(handlers.ListHistory, e))

	return historyGroup
}
