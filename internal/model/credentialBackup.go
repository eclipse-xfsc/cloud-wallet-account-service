package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"gorm.io/gorm"
	"time"
)

type ContextKey string

const InvalidateContextKey ContextKey = "invalidate"
const SkipBeforeUpdateHookContextKey ContextKey = "skipHook"

type Backup struct {
	*gorm.Model
	BindingId   sql.NullString `gorm:"binding_id"`
	UserId      string         `gorm:"user_id"`
	Name        string         `gorm:"name"`
	Credentials []byte         `gorm:"credentials"`
}

func (b *Backup) AfterFind(tx *gorm.DB) (err error) {
	invalidate := tx.Statement.Context.Value(InvalidateContextKey)
	if invalidate != nil && invalidate.(bool) {
		backup := tx.Statement.Dest.(*Backup)
		backup.BindingId = sql.NullString{String: "", Valid: false}
		return tx.WithContext(context.WithValue(context.Background(), SkipBeforeUpdateHookContextKey, true)).
			Save(backup).Error
	}
	return nil
}

func (b *Backup) BeforeUpdate(tx *gorm.DB) error {
	skip := tx.Statement.Context.Value(SkipBeforeUpdateHookContextKey)
	if skip != nil && skip.(bool) {
		return nil
	}

	tz, _ := time.LoadLocation("UTC")
	created := b.CreatedAt.In(tz)
	if time.Now().In(tz).Sub(created) > config.ServerConfiguration.BackupLinkTTL {
		tx.WithContext(context.WithValue(context.Background(), SkipBeforeUpdateHookContextKey, true)).
			Updates(Backup{BindingId: sql.NullString{String: "", Valid: false}})
		return fmt.Errorf("time for backup update expired. created %s - ttl %s", created, config.ServerConfiguration.BackupLinkTTL)
	} else {
		return nil
	}
}

func GetLastBackup(db *gorm.DB, id string) (Backup, error) {
	var backup Backup
	err := setSchema(db).
		Where("user_id=?", id).
		Order("created_at DESC").
		First(&backup).Error
	return backup, err
}

func GetBackups(db *gorm.DB, id string, after time.Time) ([]Backup, error) {
	var backups []Backup
	tbl := setSchema(db).
		Where("user_id=? AND created_at >= ?", id, after).
		Order("created_at DESC").
		Find(&backups)

	return backups, tbl.Error
}

func GetBackup(db *gorm.DB, bindingId string) ([]Backup, error) {
	var backup Backup
	err := setSchema(db).
		WithContext(context.WithValue(context.Background(), InvalidateContextKey, true)).
		Where("binding_id=?", bindingId).
		First(&backup).Error
	return []Backup{backup}, err
}

func CreateBackupDBEntry(db *gorm.DB, bidingId string, userId string, name string, creds []byte) error {
	record := Backup{
		BindingId:   sql.NullString{String: bidingId, Valid: true},
		UserId:      userId,
		Name:        name,
		Credentials: creds,
	}
	sq := setSchema(db).Create(&record)
	return sq.Error
}

func EnrichBackupDBEntry(db *gorm.DB, bidingId string, creds []byte) error {
	var record Backup
	sq := setSchema(db).Where("binding_id=?", bidingId).First(&record)
	if sq.Error != nil {
		return sq.Error
	}
	if len(record.Credentials) > 0 {
		return fmt.Errorf("cannot update immutable backup data")
	}
	return db.Model(&record).Where("binding_id=?", bidingId).Updates(Backup{Credentials: creds}).Error
}

func DeleteBackupById(db *gorm.DB, bindingId string) error {
	var backup Backup
	return setSchema(db).Where("binding_id=?", bindingId).Delete(&backup).Error
}

func DeleteInvalidatedBackups(db *gorm.DB, userId string) error {
	var backups []Backup
	return setSchema(db).Where("user_id=? AND binding_id=?", userId, sql.NullString{String: "", Valid: false}).Delete(&backups).Error
}
