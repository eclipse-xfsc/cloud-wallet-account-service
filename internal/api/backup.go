package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func CredentialBackupRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {
	backupGroup := group.Group("/backup")
	backupGroup.PUT("/:id/:bid", common.ConstructResponse(handlers.CreateBackupCredentials, e))
	backupGroup.GET("/:id/:bid", common.ConstructResponse(handlers.GetBackupCredentials, e))
	backupGroup.GET("/link/:mode", common.ConstructResponse(handlers.GenerateBackupLink, e))
	backupGroup.GET("/all", common.ConstructResponse(handlers.GetAllBackupCredentials, e))
	backupGroup.GET("/latest", common.ConstructResponse(handlers.GetLastBackupCredentials, e))
	backupGroup.DELETE("/invalid", common.ConstructResponse(handlers.DeleteInvalidUserBackups, e))
	backupGroup.DELETE("/:bid", common.ConstructResponse(handlers.DeleteBackup, e))
	return backupGroup
}
