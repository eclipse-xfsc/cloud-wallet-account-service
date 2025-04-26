package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/middleware"
)

const (
	presentationRequestEventMessage = "presentation request received over REST API"
	proofEventMessage               = "Created proof"
)

func PresentationsRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {

	presentationGroup := group.Group("/presentations")

	presentationGroup.POST("/list", common.ConstructResponse(handlers.ListPresentations, e))

	presentationGroup.GET("/selection/:id", middleware.WithHistoryRecord(common.PresentationRequest,
		presentationRequestEventMessage, e),
		common.ConstructResponse(handlers.GetPresentationRequest, e),
	)
	presentationGroup.POST("/proof/:id", middleware.WithHistoryRecord(common.Presented,
		proofEventMessage, e),
		common.ConstructResponse(handlers.CreatePresentation, e),
	)
	presentationGroup.GET("/selection/all", common.ConstructResponse(handlers.GetPresentationDefinitions, e))

	return presentationGroup
}
