package handlers

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/model"
	"gorm.io/gorm"
)

type Backup struct {
	BindingId   string    `json:"bindingId,omitempty"`
	UserId      string    `json:"user_id"`
	Name        string    `json:"name"`
	Credentials []byte    `json:"credentials"`
	Timestamp   time.Time `json:"timestamp"`
}

type GetBackupsOutput struct {
	Backups []Backup `json:"backups"`
}

type BackupLinkOutput struct {
	Path string  `json:"path"`
	TTL  float64 `json:"expiresInSeconds"`
}

func fromModelBackupToOutputBackup(from model.Backup) Backup {
	return Backup{
		BindingId: from.BindingId.String, UserId: from.UserId, Credentials: from.Credentials, Timestamp: from.CreatedAt, Name: from.Name,
	}
}

// CreateBackupCredentials godoc
// @Summary Create backup credentials
// @Description Create backup credentials for a user
// @Tags credentials
// @Accept  bytes
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "User ID"
// @Param bid path string true "Backup ID"
// @Param data body string true "Credentials data bytes"
// @Success 200 string
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/{id}/{bid} [put]

func CreateBackupCredentials(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bid := ctx.Param("bid")
	if bid == "" {
		return "", common.ErrorResponseBadRequest(ctx, "url param `bid` required", nil)
	}
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	crCtx := types.CryptoContext{
		Namespace: e.GetNamespace(),
		Group:     "account",
		Context:   context.Background(),
	}

	keyId := types.CryptoIdentifier{
		KeyId:         user.ID(),
		CryptoContext: crCtx,
	}

	var er error
	var exists bool

	if exists, er = e.GetCryptoProvider().IsCryptoContextExisting(crCtx); er == nil && !exists {
		er = e.GetCryptoProvider().CreateCryptoContext(crCtx)
	}
	if er != nil {
		return nil, er
	}

	if exists, er = e.GetCryptoProvider().IsKeyExisting(keyId); er == nil && !exists {
		er = e.GetCryptoProvider().GenerateKey(types.CryptoKeyParameter{
			Identifier: keyId,
			KeyType:    types.Aes256GCM,
		})
	}
	if er != nil {
		return nil, er
	}
	data, err = e.GetCryptoProvider().Encrypt(keyId, data)
	if err != nil {
		return nil, err
	}
	return "", model.EnrichBackupDBEntry(e.GetDB(), bid, data)
}

// GetAllBackupCredentials godoc
// @Summary Get backup credentials
// @Description Get backup credentials for a user
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param Authorization header string true "user bearer token obtained from OIDC provider"
// @Success 200 {object} GetBackupsOutput
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/all [get]

func GetAllBackupCredentials(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var after time.Time
	if tmp := ctx.Query("after"); tmp != "" {
		after, err = time.Parse(time.UnixDate, tmp)
	}
	if err != nil {
		return nil, err
	}
	backups, err := model.GetBackups(e.GetDB(), user.ID(), after)
	if err != nil {
		return nil, err
	}
	keyId := types.CryptoIdentifier{
		KeyId: user.ID(),
		CryptoContext: types.CryptoContext{
			Namespace: e.GetNamespace(),
			Group:     "account",
			Context:   context.Background(),
		},
	}
	var res = GetBackupsOutput{Backups: make([]Backup, 0)}
	for _, b := range backups {
		if b.Credentials != nil && len(b.Credentials) > 0 {
			b.Credentials, err = e.GetCryptoProvider().Decrypt(keyId, b.Credentials)
			if err != nil {
				return nil, err
			}
		}
		res.Backups = append(res.Backups, fromModelBackupToOutputBackup(b))
	}
	return res, nil
}

// GetBackupCredentials godoc
// @Summary Get exact credentials backup
// @Description Get exact credentials backup for a user
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "User ID"
// @Param bid path string true "Backup ID"
// @Success 200 {object} GetBackupsOutput
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/{id}/{bid} [get]

func GetBackupCredentials(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bid := ctx.Param("bid")
	if bid == "" {
		return "", common.ErrorResponseBadRequest(ctx, "url param `bid` required", nil)
	}

	backups, err := model.GetBackup(e.GetDB(), bid)
	if err != nil {
		return nil, err
	}
	keyId := types.CryptoIdentifier{
		KeyId: user.ID(),
		CryptoContext: types.CryptoContext{
			Namespace: e.GetNamespace(),
			Group:     "account",
			Context:   context.Background(),
		},
	}

	var res = GetBackupsOutput{Backups: make([]Backup, 0)}
	for _, b := range backups {
		if len(b.Credentials) > 0 {
			b.Credentials, err = e.GetCryptoProvider().Decrypt(keyId, b.Credentials)
			if err != nil {
				return nil, err
			}
		}
		res.Backups = append(res.Backups, fromModelBackupToOutputBackup(b))
	}
	return res, nil
}

// GetLastBackupCredentials godoc
// @Summary Get last backup credentials
// @Description Get backup credentials for a user
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param Authorization header string true "user bearer token obtained from OIDC provider"
// @Success 200 {object} GetBackupsOutput
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/all [get]

func GetLastBackupCredentials(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	backup, err := model.GetLastBackup(e.GetDB(), user.ID())
	if err != nil {
		return nil, err
	}
	keyId := types.CryptoIdentifier{
		KeyId: user.ID(),
		CryptoContext: types.CryptoContext{
			Namespace: e.GetNamespace(),
			Group:     "account",
			Context:   context.Background(),
		},
	}
	backup.Credentials, err = e.GetCryptoProvider().Decrypt(keyId, backup.Credentials)
	return GetBackupsOutput{Backups: []Backup{fromModelBackupToOutputBackup(backup)}}, nil
}

// GenerateBackupLink godoc
// @Summary Generate  or download backup credentials link
// @Description Get backup credentials link to upload or download credentials backup. The link can be used externally with no Authorization
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param Authorization header string true "user bearer token obtained from OIDC provider"
// @Param mode path string true "upload or download"
// @Param name query string false "Name of backup to upload"
// @Param bindingId query string false "id of backup to download. required if mod=download"
// @Success 200 {object} BackupLinkOutput
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/link/{mode} [get]

func GenerateBackupLink(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	mode := ctx.Param("mode")
	name := ctx.Query("name")
	var bid string
	var dbOperation func(*gorm.DB, string, string, string, []byte) error
	var prefix = "xfscpcmbackup"
	if mode == common.ModeUpload {
		bid = e.GetRandomId()
		dbOperation = model.CreateBackupDBEntry
		prefix = prefix + "upload"
	} else if mode == common.ModeDownload {
		bid = ctx.Query("bindingId")
		prefix = prefix + "download"
		if bid == "" {
			return nil, common.ErrorResponseBadRequest(ctx, "queryParam `bindingId` required", nil)
		}
		dbOperation = func(db *gorm.DB, s string, s2 string, s3 string, bytes []byte) error {
			return nil
		}
	} else {
		return nil, fmt.Errorf("unknown backup link mode")
	}

	path := getOneTimePath(ctx, bid, user)
	res := BackupLinkOutput{Path: prefix + "://" + path.String(), TTL: config.ServerConfiguration.BackupLinkTTL.Seconds()}
	return res, dbOperation(e.GetDB(), bid, user.ID(), name, []byte{})
}

func getOneTimePath(ctx *gin.Context, bid string, user *common.UserInfo) *url.URL {
	path := &url.URL{}
	if ctx.Request.TLS == nil {
		path.Scheme = "http"
	} else {
		path.Scheme = "https"
	}
	path.Host = ctx.Request.Host
	path.Path = fmt.Sprintf("api/accounts/credentials/backup/%s/%s", user.ID(), bid)
	path.RawQuery = url.Values{}.Encode()
	return path
}

// DeleteBackup godoc
// @Summary Delete backup credentials
// @Description Delete backup credentials for a user
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param Authorization header string true "user bearer token obtained from OIDC provider"
// @Param bid path string true "Backup ID"
// @Success 200 {object} GetBackupsOutput
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/{bid} [delete]

func DeleteBackup(ctx *gin.Context, e common.Env) (any, error) {
	bid := ctx.Param("bid")
	return "", model.DeleteBackupById(e.GetDB(), bid)
}

// DeleteInvalidUserBackups godoc
// @Summary Delete invalidated backup credentials
// @Description Delete invalidated backup credentials for a user
// @Tags credentials
// @Accept  json
// @Produce  json
// @Param tenantId path string true "Tenant ID"
// @Param Authorization header string true "user bearer token obtained from OIDC provider"
// @Param bid path string true "Backup ID"
// @Success 200 {object} GetBackupsOutput
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
// @Router /credentials/backup/invalid [delete]

func DeleteInvalidUserBackups(ctx *gin.Context, e common.Env) (any, error) {
	user, err := common.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return "", model.DeleteInvalidatedBackups(e.GetDB(), user.ID())
}
