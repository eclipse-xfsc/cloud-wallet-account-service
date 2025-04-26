package messaging

import (
	"context"
	"errors"
	"github.com/Nerzal/gocloak/v13"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
	"net/http"
)

type PresentationRequestRecord struct {
	messaging.HistoryRecord
	TTL int `json:"ttl"`
}

func HandlePresentationRequest(e event.Event, env common.Env) error {
	typ := common.RecordEventType(e.Type())
	if typ != common.PresentationRequest {
		return errors.ErrUnsupported
	}
	var record PresentationRequestRecord
	err := e.DataAs(&record)
	if err != nil {
		logger.Error(err, "could not unpack event", "type", typ, "id", e.ID())
		return err
	}
	ctx := &gin.Context{Request: &http.Request{}}
	ctx.Params = append(gin.Params{}, gin.Param{Key: "requestId", Value: record.RequestId})
	withValue := context.WithValue(context.Background(), common.UserKey, &common.UserInfo{UserInfo: &gocloak.UserInfo{Sub: &record.UserId}})
	withValue = context.WithValue(withValue, common.TTLKey, record.TTL)
	ctx.Request = ctx.Request.WithContext(withValue)
	var handler = func() error {
		_, er := handlers.GetPresentationRequest(ctx, env)
		return er
	}
	return handler()
}
