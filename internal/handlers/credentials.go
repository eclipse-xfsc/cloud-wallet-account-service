package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cloudeventprovider "gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/credential"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	msgCommon "gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
)

var logger = common.GetLogger()

type ListCredentialRequestBody struct {
	Search string `json:"search"`
}

// ListCredentials godoc
// @Summary List credentials
// @Description Lists all credentials
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param search body ListCredentialRequestBody false "Search"
// @Success 200 {object} []presentation.FilterResult
// @Failure 400 {object} common.ServerErrorResponse
// @Router /credentials/list [get]
func ListCredentials(ctx *gin.Context, e common.Env) (any, error) {
	stor := services.GetStorage(e.GetHttpClient())
	user, err := common.GetUserFromContext(ctx)
	auth := ctx.Request.Header.Get("Authorization")
	if err != nil {
		return nil, err
	} else {
		var requestBody ListCredentialRequestBody
		jsonData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			return nil, common.ErrorResponseBadRequest(ctx, "cannot parse the body", nil)
		}

		if len(jsonData) != 0 {
			err = json.Unmarshal(jsonData, &requestBody)
			if err != nil {
				return nil, common.ErrorResponseBadRequest(ctx, "cannot parse the json body", nil)
			}
		}
		if requestBody.Search != "" {
			constraints := buildConstraints(requestBody.Search)
			return stor.GetCredentials(auth, user.ID(), constraints)
		}
		return stor.GetCredentials(auth, user.ID(), nil)
	}
}

// ListPresentations godoc
// @Summary List presentations
// @Description Lists all presentations
// @Tags presentations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param search body ListCredentialRequestBody false "Search"
// @Success 200 {object} []presentation.FilterResult
// @Failure 400 {object} common.ServerErrorResponse
// @Router /presentations/list [get]
func ListPresentations(ctx *gin.Context, e common.Env) (any, error) {
	stor := services.GetStorage(e.GetHttpClient())
	tmp := ctx.Request.Context().Value(common.UserKey)
	auth := ctx.Request.Header.Get("Authorization")
	if user, ok := tmp.(*common.UserInfo); !ok {
		return nil, common.ErrorResponseBadRequest(ctx, "cannot extract user data from request context", nil)
	} else {
		var requestBody ListCredentialRequestBody
		jsonData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			return nil, common.ErrorResponseBadRequest(ctx, "cannot parse the body", nil)
		}

		if len(jsonData) != 0 {
			err = json.Unmarshal(jsonData, &requestBody)
			if err != nil {
				return nil, common.ErrorResponseBadRequest(ctx, "cannot parse the json body", nil)
			}
		}
		if requestBody.Search != "" {
			constraints := buildConstraints(requestBody.Search)
			return stor.GetPresentations(auth, user.ID(), constraints)
		}
		return stor.GetPresentations(auth, user.ID(), nil)
	}
}

func DeleteCredential(ctx *gin.Context, e common.Env) (any, error) {
	mock := make(map[string]string)
	return mock, nil
}

func RevokeCredential(ctx *gin.Context, e common.Env) (any, error) {
	mock := make(map[string]string)
	return mock, nil
}

func buildConstraints(searchValue string) *presentation.PresentationDefinition {
	field := presentation.Field{
		Path: []string{"$.credentialSubject"},
		Filter: &presentation.Filter{
			Pattern: searchValue,
		},
	}
	constraints := presentation.Constraints{
		LimitDisclosure: "",
		Fields:          []presentation.Field{field},
	}
	inputDescriptor := presentation.InputDescriptor{
		Format:      presentation.Format{},
		Constraints: constraints,
	}
	result := &presentation.PresentationDefinition{
		InputDescriptors: []presentation.InputDescriptor{inputDescriptor},
		Format:           presentation.Format{LDPVC: &presentation.FormatSpecification{}},
	}
	return result
}

type IssueCredentialRequestBody struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// RequestIssuance godoc
// @Summary Request issuance of a credential
// @Description Requests the issuance of a credential
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param data body IssueCredentialRequestBody true "type and credential subject"
// @Success 200 {object} credential.CredentialOffer
// @Failure 400 {object} common.ServerErrorResponse
// @Router /credentials/issue [post]
func RequestIssuance(ctx *gin.Context, e common.Env) (any, error) {

	tenantId, b := ctx.Params.Get("tenantId")

	if !b {
		ctx.AbortWithStatus(400)
		return nil, nil
	}

	var payload map[string]interface{}
	err := json.NewDecoder(ctx.Request.Body).Decode(&payload)

	if err != nil {
		ctx.AbortWithStatus(400)
		return nil, err
	}

	t, ok := payload["type"].(string)
	p, ok2 := payload["payload"].(map[string]interface{})

	if err != nil || !ok || !ok2 {
		ctx.AbortWithStatus(400)
		return nil, err
	}

	result, err := GetCredentialConfigurations(ctx, e)

	if err != nil || !ok || !ok2 {
		ctx.AbortWithStatus(400)
		return nil, err
	}

	item, ok := result.(map[string]credential.CredentialConfiguration)[t]

	if !ok {
		ctx.AbortWithStatus(400)
		return nil, err
	}

	config := config.ServerConfiguration

	var req = messaging.IssuanceRequest{
		Request: msgCommon.Request{
			TenantId:  tenantId,
			RequestId: uuid.NewString(),
		},
		Payload:    p,
		Identifier: t,
	}

	data, err := json.Marshal(req)
	if err != nil {
		ctx.AbortWithStatus(400)
		return nil, err
	}

	logger.Debug(string(data))

	client, err := cloudeventprovider.New(
		cloudeventprovider.Config{
			Protocol: cloudeventprovider.ProtocolTypeNats,
			Settings: cloudeventprovider.NatsConfig{
				Url:          config.Nats.Url,
				QueueGroup:   config.Nats.QueueGroup,
				TimeoutInSec: config.Nats.TimeoutInSec,
			},
		},
		cloudeventprovider.ConnectionTypeReq,
		item.Subject+".request",
	)

	if err != nil {
		logger.Error(err, "error during client creation")
		ctx.AbortWithStatus(400)
		return nil, err
	}

	event, err := cloudeventprovider.NewEvent("issuance-client", messaging.EventTypeGetIssuerMetadata, data)
	if err != nil {
		ctx.AbortWithStatus(400)
		return nil, err
	}
	logger.Debug(string(data))
	repl, err := client.RequestCtx(ctx, event)
	if err != nil || repl == nil {
		logger.Error(err, "error during issue request")
		ctx.AbortWithStatus(400)
		return nil, err
	}
	var metadata messaging.IssuanceReply
	err = json.Unmarshal(repl.DataEncoded, &metadata)
	logger.Debug(string(repl.DataEncoded))
	if err != nil || metadata.Error != nil {
		if err == nil {
			logger.Error(errors.New("issuance reply error"), metadata.Error.Msg)
		} else {
			logger.Error(err, "Error during issue reply")
		}

		ctx.AbortWithStatus(400)
		return nil, err
	}

	return metadata.Offer, nil
}

// GetCredentialConfigurations godoc
// @Summary Get credential configurations
// @Description Retrieves all credential configurations
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} map[string]credential.CredentialConfiguration
// @Failure 400 {object} common.ServerErrorResponse
// @Router /credentials/schemas [get]
func GetCredentialConfigurations(ctx *gin.Context, e common.Env) (any, error) {

	tenantId, b := ctx.Params.Get("tenantId")

	if !b {
		logger.Info("no tenant id found")
		ctx.AbortWithStatus(400)
		return nil, nil
	}

	config := config.ServerConfiguration

	var req = messaging.GetIssuerMetadataReq{
		Request: msgCommon.Request{
			TenantId:  tenantId,
			RequestId: uuid.NewString(),
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		logger.Error(err, "no tenant id found")
		ctx.AbortWithStatus(400)
		return nil, err
	}

	client, err := cloudeventprovider.New(
		cloudeventprovider.Config{
			Protocol: cloudeventprovider.ProtocolTypeNats,
			Settings: cloudeventprovider.NatsConfig{
				Url:          config.Nats.Url,
				QueueGroup:   config.Nats.QueueGroup,
				TimeoutInSec: time.Second * 10,
			},
		},
		cloudeventprovider.ConnectionTypeReq,
		messaging.TopicGetIssuerMetadata,
	)

	if err != nil {
		logger.Error(err, "no error creating client")
		ctx.AbortWithStatus(400)
		return nil, err
	}

	event, err := cloudeventprovider.NewEvent("metadata-client", messaging.EventTypeGetIssuerMetadata, data)
	if err != nil {
		logger.Error(err, "error creating event")
		ctx.AbortWithStatus(400)
		return nil, err
	}

	repl, err := client.RequestCtx(ctx, event)
	if err != nil || repl == nil {
		logger.Error(err, "error during nats request")
		ctx.AbortWithStatus(400)
		return nil, err
	}
	var metadata messaging.GetIssuerMetadataReply
	err = json.Unmarshal(repl.DataEncoded, &metadata)

	if err != nil {
		logger.Error(err, "error unmarshal reply")
		ctx.AbortWithStatus(400)
		return nil, err
	}

	return metadata.Issuer.CredentialConfigurationsSupported, nil
}
