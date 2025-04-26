package messaging

import (
	"errors"
	"github.com/cloudevents/sdk-go/v2/event"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"slices"
)

var logger = common.GetLogger()

func HandleHistoryRecord(e event.Event, env common.Env) error {
	typ := common.RecordEventType(e.Type())
	if !slices.Contains(common.RecordEventTypes(), typ) {
		return errors.ErrUnsupported
	}
	var record messaging.HistoryRecord
	err := e.DataAs(&record)
	if err != nil {
		logger.Error(err, "could not unpack event", "type", typ, "id", e.ID())
		return err
	}
	err = model.CreateRecordDBEntry(env.GetDB(), record.UserId, typ, record.Message)
	if err != nil {
		logger.Error(err, "could not write record to db", "type", typ, "id", e.ID())
		return err
	}
	return nil
}
