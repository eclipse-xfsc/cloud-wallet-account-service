package model

import (
	"fmt"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"gorm.io/gorm"
)

const tableSchema = "accounts"

var logger = common.GetLogger()

func setSchema(db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}
	return db.Exec(fmt.Sprintf("SET search_path TO %s", tableSchema))
}
