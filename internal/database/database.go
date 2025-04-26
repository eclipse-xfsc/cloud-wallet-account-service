package database

import (
	"fmt"
	"time"

	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gLog "gorm.io/gorm/logger"
)

var logger = common.GetLogger()

func NewDatabaseConnection(databaseType common.DatabaseType) (*gorm.DB, error) {
	switch databaseType {
	case common.Postgres:
		db, err := newPsgConnection()
		if err != nil {
			return nil, err
		}
		return db, nil
	default:
		return nil, fmt.Errorf("unknown databaseType: %s. Could not create Database Connection", databaseType)
	}
}

func newPsgConnection() (*gorm.DB, error) {
	dbConfig := config.ServerConfiguration.Database

	dsn := fmt.Sprintf(
		"host=%s port=%v user=%s password=%s dbname=%s sslmode=disable TimeZone=CET",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gLog.Default.LogMode(gLog.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, err
	}

	logger.Info("Postgress db", db.Name(), " connected")

	databaseMigrations(db)

	return db, nil
}
