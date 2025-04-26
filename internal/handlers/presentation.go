package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
	"io"
	"net/http"
)

type CreateProofPayload struct {
	SignKeyId string `json:"signKeyId"`
	Filters   []presentation.FilterResult
}

// GetPresentationRequest godoc
// @Summary Get presentation request
// @Description Retrieves a presentation request
// @Tags presentations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "Presentation ID"
// @Success 200 {object} []presentation.FilterResult
// @Failure 400 {object} common.ServerErrorResponse
// @Router /presentations/selection/{id} [get]
func GetPresentationRequest(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	ttl, ok := ctx.Request.Context().Value(common.TTLKey).(int)
	if !ok {
		ttl = 0
	}

	credentialVerification := services.GetCredentialVerification(e.GetHttpClient())
	request, err := getPresentationRequest(ctx, credentialVerification)
	if err != nil {
		return nil, err
	}
	err = model.CreatePresentationRequestDBEntry(e.GetDB(), user.ID(), request.Id, ttl, request.RequestId)

	if err == nil {
		err = credentialVerification.AssignProof(request.Id, user.ID())
		if err != nil {
			return nil, err
		}
		err = credentialVerification.AssignProof(request.RequestId, user.ID())
		if err != nil {
			return nil, err
		}
	} else if !errors.Is(err, model.PresentationAlreadyExistsError) {
		return nil, err
	} else {
	}
	return services.GetStorage(e.GetHttpClient()).GetCredentials("", user.ID(), &request.PresentationDefinition)
}

func getPresentationRequest(ctx *gin.Context, presentationService *services.CredentialVerification) (*services.PresentationRequest, error) {
	presentationId := ctx.Param("id")
	withId := presentationId != ""
	presentationRequestId := ctx.Param("requestId")
	withRequestId := presentationRequestId != ""
	if withId && withRequestId {
		return nil, fmt.Errorf("either `id` or `requestId` query params must be provided. Not both")
	}
	if withId {
		return presentationService.GetProofRequest(presentationId)
	} else if withRequestId {
		return presentationService.GetProofRequestByProofRequestId(presentationRequestId)
	} else {
		return nil, fmt.Errorf("no `id` or `requestId` query params provided")
	}
}

// CreatePresentation godoc
// @Summary Create a presentation
// @Description Creates a presentation
// @Tags presentations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "Presentation ID"
// @Param data body CreateProofPayload true "Proof payload"
// @Success 200 {object} nil
// @Failure 400 {object} common.ServerErrorResponse
// @Router /presentations/proof/{id} [post]
func CreatePresentation(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	presentationId := ctx.Param("id")
	if presentationId == "" {
		return nil, fmt.Errorf("id of presentationRequest not provided")
	}
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	var proofPayload CreateProofPayload
	err = json.Unmarshal(data, &proofPayload)
	if err != nil {
		return nil, err
	}
	signer := services.GetSigner(e.GetHttpClient())
	// fetch list of did docs to match the did doc id that matches passed from client
	didObj, err := signer.ListDidDocs(e.GetNamespace(), user.ID())
	if err != nil {
		return nil, err
	}
	var didId = ""
	for _, did := range didObj.List {
		if did.Name == proofPayload.SignKeyId {
			didId = did.Did
		}
	}
	if didId == "" {
		return nil, common.ErrorResponse(ctx, http.StatusNotFound, "did not find DID for provided `signKeyId`", nil)
	}
	presRequest, err := model.GetPresentationRequestById(e.GetDB(), user.ID(), presentationId)
	if err != nil {
		return nil, common.ErrorResponse(ctx, http.StatusNotFound, fmt.Sprintf("did not find presentation request with id `%s`", presentationId), nil)
	}
	err = services.
		GetCredentialVerification(e.GetHttpClient()).
		CreateProof(presRequest.ProofRequestId, proofPayload.Filters, e.GetNamespace(), user.ID(), proofPayload.SignKeyId, didId)
	if err == nil {
		err = model.DeletePresentationRequests(e.GetDB(), user.ID(), []string{presentationId})
	}
	return nil, err
}

// GetPresentationDefinitions godoc
// @Summary Get presentation definitions
// @Description Retrieves all presentation definitions
// @Tags presentations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} []presentation.PresentationDefinition
// @Failure 400 {object} common.ServerErrorResponse
// @Router /presentations/selection/all [get]
func GetPresentationDefinitions(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	presentations, err := model.GetAllPresentationRequests(e.GetDB(), user.ID())
	if err != nil {
		return nil, err
	}
	var res = []presentation.PresentationDefinition{}
	for _, pres := range presentations {
		request, err := services.GetCredentialVerification(e.GetHttpClient()).GetProofRequest(pres.RequestId)
		if err != nil {
			logger.Error(err, "could not get proof request", "id", pres.RequestId)
			return nil, err
		}
		request.PresentationDefinition.Id = pres.RequestId
		res = append(res, request.PresentationDefinition)
	}
	return res, nil
}
