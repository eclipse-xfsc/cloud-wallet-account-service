package messaging

import (
	"bytes"
	"encoding/json"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/handlers"
	"io"
	"net/http"
)

const CreateKeyEventType = "createKey"

type CreateKeyEventData struct {
	messaging.HistoryRecord
	handlers.CreateDidPayload
}

func HandleCreateKey(e event.Event, env common.Env) error {
	if e.Type() != CreateKeyEventType {
		return nil
	}
	var data CreateKeyEventData
	err := e.DataAs(&data)
	if err != nil {
		return err
	}
	createKeyPayloadBytes, err := json.Marshal(data.CreateDidPayload)
	if err != nil {
		return err
	}
	ctx := &gin.Context{}
	ctx.Request = &http.Request{Body: io.NopCloser(bytes.NewBuffer(createKeyPayloadBytes))}
	ctx = common.GetContextWithUserId(ctx, data.UserId)
	return WrapEndpointHandler(handlers.CreateDID, ctx)(e, env)
}
