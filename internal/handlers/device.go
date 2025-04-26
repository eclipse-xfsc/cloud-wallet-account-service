package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	cmn "github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
)

const (
	invitationEventType = "device.connection"
	invitationTopic     = "remotecontrol.xxx"
	invitationProtocol  = "nats"
)

// ListDevices godoc
// @Summary List devices
// @Description Lists all devices
// @Tags device
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} []services.DIDCommConnection
// @Failure 400 {object} common.ServerErrorResponse
// @Router /devices/list [get]
func ListDevices(ctx *gin.Context, e cmn.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	didcomm := services.GetDIDComm(nil)

	list, err := didcomm.GetConnectionList(user.ID(), ctx.Query("search"))

	if err != nil {
		return nil, err
	}

	return list, nil
}

// LinkDevice godoc
// @Summary Link a device
// @Description Links a device
// @Tags device
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {string} string "Link"
// @Router /devices/link [get]
func LinkDevice(ctx *gin.Context, e cmn.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	didcomm := services.GetDIDComm(nil)

	reqBody := services.InvitationRequestBody{
		Protocol:  invitationProtocol,
		Topic:     invitationTopic,
		Group:     user.ID(),
		EventType: invitationEventType,
		Properties: map[string]string{
			"account":  user.ID(),
			"greeting": "hello-world",
		},
	}

	link, err := didcomm.GetInviteLink(reqBody)
	if err != nil {
		return nil, err
	}

	response := make(map[string]interface{})
	response["qrCodeLink"] = link

	return link, nil
}

// DeleteDevice godoc
// @Summary Delete a device
// @Description Deletes a device
// @Tags device
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "Device ID"
// @Success 200 "Success"
// @Failure 400 {object} common.ServerErrorResponse
// @Router /devices/{id} [delete]
func DeleteDevice(ctx *gin.Context, e cmn.Env) (any, error) {

	did := ctx.Param("id")
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	db := e.GetDB()

	_, err = model.GetUserConnectionDbEntry(db, user.ID(), did)

	if err != nil {
		return nil, err
	}

	didcomm := services.GetDIDComm(nil)

	err = didcomm.DeleteConnection(did)
	if err != nil {
		return nil, err
	}

	err = model.DeleteUserConnectioniDbEntry(db, user.ID(), did)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BlockDevice godoc
// @Summary Block a device
// @Description Blocks a device
// @Tags device
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "Device ID"
// @Success 200 "Success"
// @Router /devices/block/{id} [post]
func BlockDevice(ctx *gin.Context, e cmn.Env) (any, error) {
	did := ctx.Param("id")
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	db := e.GetDB()

	_, err = model.GetUserConnectionDbEntry(db, user.ID(), did)

	if err != nil {
		return nil, err
	}

	didcomm := services.GetDIDComm(nil)

	err = didcomm.BlockConnection(did)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func AddDevice(ctx *gin.Context, e cmn.Env) (any, error) {
	mock := make(map[string]string)
	mock["name"] = "My Device"
	mock["connectionId"] = "4729423498hsdfndfj3jsjsj"
	mock["pairingDate"] = "12.12.23"
	mock["qr"] = "aGVsbG8gd29ybGQ="
	return mock, nil
}
