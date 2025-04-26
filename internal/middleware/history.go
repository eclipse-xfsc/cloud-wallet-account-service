package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"net/http"
)

func WithHistoryRecord(event common.RecordEventType, msg string, env common.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := common.GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"message": common.HistoryRecordError})
			c.Abort()
			return
		}
		err = model.CreateRecordDBEntry(env.GetDB(), user.ID(), event, msg)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"message": common.HistoryRecordError})
			c.Abort()
			return
		}
	}
}
