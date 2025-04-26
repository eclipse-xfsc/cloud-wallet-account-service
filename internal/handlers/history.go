package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"time"
)

type Event struct {
	Message   string                 `json:"event"`
	EventType common.RecordEventType `json:"type"`
	UserId    string                 `json:"userId"`
	Timestamp time.Time              `json:"timestamp"`
}

type ListHistoryOutput struct {
	Events []Event `json:"events"`
}

func fromRecordToHistoryElement(record model.HistoryRecord) Event {
	el := Event{
		UserId:    record.UserId,
		EventType: record.EventType,
		Timestamp: record.CreatedAt,
		Message:   record.Message,
	}
	return el
}

// ListHistory godoc
// @Summary List history
// @Description Lists all history events
// @Tags history
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} ListHistoryOutput
// @Failure 400 {object} common.ServerErrorResponse
// @Router /history/list [get]
func ListHistory(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	records, err := model.GetRecords(e.GetDB(), user.ID())
	if err != nil {
		return nil, err
	}
	var history = ListHistoryOutput{Events: make([]Event, 0)}
	for _, rec := range records {
		history.Events = append(history.Events, fromRecordToHistoryElement(rec))
	}
	return history, nil
}
