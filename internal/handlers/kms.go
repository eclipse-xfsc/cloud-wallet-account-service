package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
)

type DID struct {
	Id        string    `json:"id"`
	Did       string    `json:"did"`
	Detail    string    `json:"detail,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type ListDIDResponse struct {
	List []DID `json:"list"`
}

type CreateDidPayload struct {
	KeyType types.KeyType `json:"keyType"`
}

func (p CreateDidPayload) validate() bool {
	return types.ValidateMethod(p.KeyType)
}

func fromSignerDidToKMSDid(from services.ListDidItem) DID {
	return DID{
		Id:  from.Name,
		Did: from.Did,
	}
}

// GetSupportedKeysAlgs godoc
// @Summary Get supported keys algorithms
// @Description Retrieves the supported keys algorithms
// @Tags kms
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Failure 400 {object} common.ServerErrorResponse
// @Success 200 {array} types.KeyType
// @Router /kms/keyTypes [get]
func GetSupportedKeysAlgs(ctx *gin.Context, e common.Env) (any, error) {
	//return e.GetCryptoProvider().GetSupportedKeysAlgs(), nil
	// todo we need to filter ListDID by supported keyType. Until then only es256 keys are supported
	return []types.KeyType{types.Ed25519}, nil
}

// ListDID godoc
// @Summary List all DIDs
// @Description Lists all DIDs
// @Tags kms
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Failure 400 {object} common.ServerErrorResponse
// @Success 200 {object} ListDIDResponse
// @Router /kms/did/list [get]
func ListDID(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	dids, err := services.GetSigner(e.GetHttpClient()).ListDidDocs(e.GetNamespace(), user.ID())
	if err != nil {
		return nil, err
	}
	var res = ListDIDResponse{List: make([]DID, 0)}
	for _, did := range dids.List {
		res.List = append(res.List, fromSignerDidToKMSDid(did))
	}
	return res, nil
}

// CreateDID godoc
// @Summary Create a DID
// @Description Creates a DID
// @Tags kms
// @Accept  json
// @Produce  json
// @Param data body CreateDidPayload true "DID payload"
// @Failure 400 {object} common.ServerErrorResponse
// @Success 200 {string} string "DID ID"
// @Router /kms/did/create [post]
func CreateDID(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	var payload CreateDidPayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	if !payload.validate() {
		return nil, fmt.Errorf("invalid key type %s", payload.KeyType)
	}

	cryptoCtx := types.CryptoContext{
		Namespace: e.GetNamespace(),
		Group:     user.ID(),
		Context:   context.Background(),
	}
	var exists bool
	var er error
	if exists, er = e.GetCryptoProvider().IsCryptoContextExisting(cryptoCtx); er == nil && !exists {
		er = e.GetCryptoProvider().CreateCryptoContext(cryptoCtx)
	}
	if er != nil {
		return nil, er
	}

	cryptoId := types.CryptoIdentifier{
		KeyId:         e.GetRandomId(),
		CryptoContext: cryptoCtx,
	}

	cryptoParam := types.CryptoKeyParameter{
		Identifier: cryptoId,
		KeyType:    payload.KeyType,
	}
	err = e.GetCryptoProvider().GenerateKey(cryptoParam)
	if err != nil {
		return nil, err
	}
	return cryptoId.KeyId, nil
}

// DeleteDID godoc
// @Summary Delete a DID
// @Description Deletes a DID
// @Tags kms
// @Accept  json
// @Produce  json
// @Param kid path string true "DID ID"
// @Failure 400 {object} common.ServerErrorResponse
// @Success 200 "Success"
// @Router /kms/did/{kid} [delete]
func DeleteDID(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	keyId := ctx.Param("kid")
	cryptoCtx := types.CryptoContext{
		Namespace: e.GetNamespace(),
		Group:     user.ID(),
		Context:   context.Background(),
	}
	var exists bool
	var er error
	if exists, er = e.GetCryptoProvider().IsCryptoContextExisting(cryptoCtx); er == nil && !exists {
		return nil,
			common.ErrorResponse(ctx,
				http.StatusNotFound,
				"provided context does not exist",
				nil)
	}
	if er != nil {
		return nil, er
	}
	cryptoId := types.CryptoIdentifier{
		KeyId:         keyId,
		CryptoContext: cryptoCtx,
	}
	return nil, e.GetCryptoProvider().DeleteKey(cryptoId)
}
