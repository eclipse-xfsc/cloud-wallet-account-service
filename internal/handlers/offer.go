package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	msgCommon "gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
	"io"
	"slices"
)

type AcceptRejectOfferPayload struct {
	KeyId string `json:"keyId"`
}

// GetCredentialOffers godoc
// @Summary Credential Offer Routes
// @Description Routes for handling credential offers
// @Tags credentials
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} []services.CredentialOffer
// @Router /credentials/offers/list [get]
func GetCredentialOffers(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	offers, err := services.GetCredentialRetrieval(e.GetHttpClient()).GetOffers(user.ID())
	if err != nil {
		slices.Reverse(*offers)
	}
	return offers, err
}

// CreateCredentialOffer godoc
// @Summary Create a credential offer
// @Description Create a credential offer
// @Tags credentials
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 "ID of created offering request"
// @Router /credentials/offers/create [post]
func CreateCredentialOffer(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	var offer services.CredentialOfferPayload
	err = json.Unmarshal(data, &offer)
	if err != nil {
		return nil, err
	}
	return services.GetCredentialRetrieval(e.GetHttpClient()).CreateOffer(user.ID(), offer)
}

// AcceptCredentialOffer godoc
// @Summary Accept a credential offer
// @Description Accept a credential offer
// @Tags credentials
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "Offer ID"
// @Success 200 "Success"
// @Router /credentials/offers/accept/{id} [post]
func AcceptCredentialOffer(ctx *gin.Context, e common.Env) (any, error) {
	return resolveCredentialOffer(ctx, e, true)
}

// DenyCredentialOffer godoc
// @Summary Deny a credential offer
// @Description Deny a credential offer
// @Tags credentials
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "Offer ID"
// @Success 200 "Success"
// @Router /credentials/offers/deny/{id} [post]
func DenyCredentialOffer(ctx *gin.Context, e common.Env) (any, error) {
	return resolveCredentialOffer(ctx, e, false)
}

func resolveCredentialOffer(ctx *gin.Context, e common.Env, accept bool) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	requestId := ctx.Param("id")

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	var payload AcceptRejectOfferPayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	var accData = services.OfferAcceptanceData{
		Accept:          accept,
		HolderKey:       payload.KeyId,
		HolderNamespace: e.GetNamespace(),
		HolderGroup:     user.ID(),
	}
	return services.GetCredentialRetrieval(e.GetHttpClient()).AcceptOffer(user.ID(), requestId, accData)
	//tenantId := ctx.Param("tenantId")
	//return publishOfferAcceptance(accept, tenantId, requestId, user, payload, e)
}

func publishOfferAcceptance(accept bool, tenantId string, requestId string, user *common.UserInfo, payload AcceptRejectOfferPayload, e common.Env) (any, error) {
	var eventMsg string
	if accept {
		eventMsg = "Credential offer accepted"
	} else {
		eventMsg = "Credential offer rejected"
	}

	acceptance := messaging.RetrievalAcceptanceNotification{
		Request: msgCommon.Request{
			TenantId:  tenantId,
			RequestId: requestId,
			GroupId:   user.ID(),
		},
		OfferingId:      requestId,
		Message:         eventMsg,
		Result:          accept,
		HolderKey:       payload.KeyId,
		HolderNamespace: e.GetNamespace(),
		HolderGroup:     user.ID(),
	}
	eventData, err := json.Marshal(acceptance)
	if err != nil {
		return nil, err
	}
	accEvent, err := cloudeventprovider.NewEvent(config.ServerConfiguration.Name, common.EventTypeOfferingAcceptance, eventData)
	if err != nil {
		return nil, err
	}
	return nil, e.AddBrokerPublication(config.ServerConfiguration.CredentialRetrival.OfferTopic, accEvent)
}
