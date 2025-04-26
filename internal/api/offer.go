package api

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
)

func CredentialOfferRoutes(group *gin.RouterGroup, e common.Env) *gin.RouterGroup {
	offerGroup := group.Group("/offers")
	offerGroup.PUT("/create", common.ConstructResponse(handlers.CreateCredentialOffer, e))
	offerGroup.GET("/list", common.ConstructResponse(handlers.GetCredentialOffers, e))
	offerGroup.POST("/:id/accept", common.ConstructResponse(handlers.AcceptCredentialOffer, e))
	offerGroup.POST("/:id/deny", common.ConstructResponse(handlers.DenyCredentialOffer, e))
	return offerGroup
}
