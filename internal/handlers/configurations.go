package handlers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
)

type SaveConfigrationsRequest struct {
	Language     string `json:"language" binding:"required"`
	HistoryLimit int    `json:"historyLimit" binding:"required"`
}

// GetUserInfo godoc
// @Summary Get user information
// @Description Retrieves user information
// @Tags configurations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} common.UserInfo
// @Failure 400 {object} common.ServerErrorResponse
// @Router /configurations/getUserInfo [get]
func GetUserInfo(ctx *gin.Context, e common.Env) (any, error) {
	tmp := ctx.Request.Context().Value(common.UserKey)
	user, ok := tmp.(*common.UserInfo)
	if !ok {
		return nil, common.ErrorResponseBadRequest(ctx, "cannot extract user data from request context", nil)
	}
	return user, nil
}

// GetConfigurations godoc
// @Summary Get configurations
// @Description Retrieves configurations
// @Tags configurations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} common.ServerErrorResponse
// @Router /configurations/list [get]
func GetConfigurations(ctx *gin.Context, e common.Env) (any, error) {
	tmp := ctx.Request.Context().Value(common.UserKey)
	user, ok := tmp.(*common.UserInfo)
	if !ok {
		return nil, common.ErrorResponseBadRequest(ctx, "cannot extract user data from request context", nil)
	}

	db := e.GetDB()

	config, err := model.GetConfigByUserID(db, user.ID())
	if err != nil {
		return nil, common.ErrorResponseBadRequest(ctx, fmt.Sprintf("cannot extract user, %s", err), nil)
	}
	return config.Attributes, nil
}

// SaveConfigurations godoc
// @Summary Save configurations
// @Description Saves configurations
// @Tags configurations
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param data body SaveConfigrationsRequest true "Configurations"
// @Success 200 {object} nil
// @Failure 400 {object} common.ServerErrorResponse
// @Router /configurations/save [post]
func SaveConfigurations(ctx *gin.Context, e common.Env) (any, error) {
	tmp := ctx.Request.Context().Value(common.UserKey)
	user, ok := tmp.(*common.UserInfo)
	if !ok {
		return nil, common.ErrorResponseBadRequest(ctx, "cannot extract user data from request context", nil)
	}

	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, common.ErrorResponseBadRequest(ctx, fmt.Sprintf("cannot parse the request, %s", err), nil)
	}

	var req SaveConfigrationsRequest
	err = json.Unmarshal(jsonData, &req)
	if err != nil {
		return nil, common.ErrorResponseBadRequest(ctx, fmt.Sprintf("cannot parse the request, %s", err), nil)
	}

	attributes := map[string]interface{}{
		"language":     req.Language,
		"historyLimit": req.HistoryLimit,
	}
	db := e.GetDB()
	err = model.CreateOrUpdateUserConfigDbEntry(db, user.ID(), attributes)
	if err != nil {
		return nil, common.ErrorResponseBadRequest(ctx, fmt.Sprintf("cannot extract user, %s", err), nil)
	}
	return nil, nil
}
