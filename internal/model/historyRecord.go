package model

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"gorm.io/gorm"
)

type HistoryRecord struct {
	*gorm.Model
	UserId    string                 `gorm:"user_id"`
	EventType common.RecordEventType `gorm:"event_type"`
	Message   string                 `gorm:"message"`
}

func GetRecords(db *gorm.DB, userId string) ([]HistoryRecord, error) {
	var records []HistoryRecord
	err := setSchema(db).Where("user_id=?", userId).Find(&records).Error
	return records, err
}

func CreateRecordDBEntry(db *gorm.DB, userId string, event common.RecordEventType, msg string) error {
	record := HistoryRecord{UserId: userId, EventType: event, Message: msg}
	sq := setSchema(db).Create(&record)
	return sq.Error
}

// WithRecord Decorator function to execute arbitrary function with populate a database with history record
func WithRecord(f common.EndpointHandler, ctx *gin.Context, env common.Env, ev common.RecordEventType, msg string) (any, error) {
	d, e := f(ctx, env)
	if e != nil {
		return nil, e
	}
	user, e := common.GetUserFromContext(ctx)
	if e != nil {
		return nil, e
	}
	e = CreateRecordDBEntry(env.GetDB(), user.ID(), ev, msg)
	return d, e
}
