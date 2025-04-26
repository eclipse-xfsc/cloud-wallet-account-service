package messaging

import (
	"errors"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
)

func HandleError(f common.EventHandler, e event.Event, env common.Env, onError func(err error)) {
	err := f(e, env)
	if err != nil && !errors.Is(err, errors.ErrUnsupported) {
		if onError != nil {
			onError(err)
		}
	}
}

func WrapEndpointHandler(from common.EndpointHandler, ctx *gin.Context) common.EventHandler {
	if ctx == nil {
		ctx = &gin.Context{}
	}
	return func(e event.Event, env common.Env) error {
		_, err := from(ctx, env)
		return err
	}
}
