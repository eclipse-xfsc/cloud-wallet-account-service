package model

import (
	"gorm.io/gorm"
)

type UserConnection struct {
	*gorm.Model
	UserId    string `gorm:"user_id"`
	RemoteDid string `gorm:"remote_did"`
}

func GetUserConnectionDbEntry(db *gorm.DB, id string, remoteDid string) (*UserConnection, error) {
	var connection UserConnection
	sq := setSchema(db).Where("user_id=? and remote_did=?", id, remoteDid).First(&connection)
	return &connection, sq.Error
}

func CreateUserConnectioniDbEntry(db *gorm.DB, id string, remoteDid string) error {
	var connection = UserConnection{RemoteDid: remoteDid, UserId: id}
	sq := setSchema(db).Create(&connection)
	return sq.Error
}

func DeleteUserConnectioniDbEntry(db *gorm.DB, id string, remoteDid string) error {
	sq := setSchema(db).Where("user_id=? and remote_did=?", id, remoteDid).Delete(&UserConnection{})
	return sq.Error
}
