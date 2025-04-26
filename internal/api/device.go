package api

import (
	"github.com/gin-gonic/gin"
	cmn "github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func DeviceRoutes(group *gin.RouterGroup, e cmn.Env) *gin.RouterGroup {

	devicesGroup := group.Group("/devices")

	devicesGroup.GET("/list", cmn.ConstructResponse(handlers.ListDevices, e))

	devicesGroup.GET("/link", cmn.ConstructResponse(handlers.LinkDevice, e))

	devicesGroup.DELETE("/:id", cmn.ConstructResponse(handlers.DeleteDevice, e))
	devicesGroup.POST("/block/:id", cmn.ConstructResponse(handlers.BlockDevice, e))

	return devicesGroup
}
