package database

import (
	"gorm.io/gorm"
)

func databaseMigrations(db *gorm.DB) {
	//  setup db migrations here
	logger.Info("Database migrations")
}
