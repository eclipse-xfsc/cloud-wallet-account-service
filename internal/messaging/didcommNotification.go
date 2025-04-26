package messaging

import (
	"errors"
	"fmt"
	"slices"

	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"

	"github.com/cloudevents/sdk-go/v2/event"
)

type DIDCommNotification struct {
	Account   string `json:"account"`
	RemoteDID string `json:"remoteDid"`
}

func HandleDIDCommNotification(e event.Event, env common.Env) error {
	fmt.Println(string(e.Data()))

	typ := common.RecordEventType(e.Type())
	if !slices.Contains(common.RecordEventTypes(), typ) {
		logger.Error(errors.ErrUnsupported, fmt.Sprintf("record type is unsupported, %s", typ))
		return errors.ErrUnsupported
	}

	var notification DIDCommNotification
	err := e.DataAs(&notification)
	if err != nil {
		logger.Error(err, "could not unpack event", "type", typ, "id", e.ID())
		return err
	}

	err = model.CreateUserConnectioniDbEntry(env.GetDB(), notification.Account, notification.RemoteDID)
	if err != nil {
		logger.Error(err, "could not write record to db", "type", typ, "id", e.ID())
		return err
	}

	return nil
}
