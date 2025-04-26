package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type UserSecret struct {
	ID        uint `gorm:"autoIncrement; not null; unique; index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	// *gorm.Model is not embedded here to avoid colliding with UserId field
	UserId   string `gorm:"user_id; unique; not null; primaryKey"`
	SecretId string `gorm:"secret_id"`
}

func GetUserSecretIdDbEntry(db *gorm.DB, id string) *UserSecret {
	var secret UserSecret
	setSchema(db).Where("user_id=?", id).First(&secret)
	return &secret
}

func CreateUserSecretDbEntry(db *gorm.DB, id string, secretId string) {
	var secret = UserSecret{SecretId: secretId, UserId: id}
	setSchema(db).Create(&secret)
}

func CreateUserSecretId(userId string) string {
	return uuid.New().String()
}
