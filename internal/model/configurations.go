package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JSONB map[string]interface{}

func (jsonField JSONB) Value() (driver.Value, error) {
	return json.Marshal(jsonField)
}

func (jsonField *JSONB) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &jsonField)
}

type UserConfig struct {
	gorm.Model
	UserID     string `gorm:"user_id"`
	Attributes JSONB  `gorm:"type:json;not null;default:'{}'"`
}

func (u *UserConfig) CreateOrUpdateConfig(db *gorm.DB) error {
	return setSchema(db).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"attributes", "updated_at"}),
			}).
		Create(u).Error
}

func CreateOrUpdateUserConfigDbEntry(db *gorm.DB, userId string, attributes map[string]interface{}) error {
	var uc = &UserConfig{UserID: userId, Attributes: attributes}
	return setSchema(db).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"attributes", "updated_at"}),
			}).
		Create(uc).Error
}

func GetConfigByUserID(db *gorm.DB, userID string) (*UserConfig, error) {
	var userConfig UserConfig
	if err := setSchema(db).Where("user_id = ?", userID).First(&userConfig).Error; err != nil {
		return nil, err
	}
	return &userConfig, nil
}
