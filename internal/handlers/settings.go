package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
)

func GetUiSettings(ctx *gin.Context, e common.Env) (any, error) {
	mock := make(map[string]string)

	return mock, nil
}

func SetUiSettings(ctx *gin.Context, e common.Env) (any, error) {
	mock := make(map[string]string)

	return mock, nil
}
