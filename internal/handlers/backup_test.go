package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type ValidBackupPayload struct {
}

func (p ValidBackupPayload) Match(v driver.Value) bool {
	return len(v.([]byte)) > 0
}

func init() {
	gin.SetMode(gin.TestMode)
	config.ServerConfiguration.KeyCloak.ExcludeEndpoints = "/backup/:id"
	config.ServerConfiguration.BackupLinkTTL = time.Minute * 5
}

func testGenerateBackupLink(mode string, bid string, name string) (any, error) {
	testUserId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Params = append(ctx.Params, gin.Param{Key: "mode", Value: mode})
	var query string
	if mode == common.ModeDownload {
		query = "?bindingId=" + bid
	} else if mode == common.ModeUpload {
		query = "?name=" + name
	} else {
		query = ""
	}
	req, _ := http.NewRequest(http.MethodGet, "https://0.0.0.0/credentials/backup/link/"+mode+query, nil)
	user := common.UserInfo{UserInfo: &gocloak.UserInfo{Sub: &testUserId}}
	req = req.WithContext(context.WithValue(req.Context(), common.UserKey, &user))
	req.TLS = &tls.ConnectionState{}
	ctx.Request = req
	e := &test.EnvObjMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectBegin()
	dbPatcher.
		ExpectQuery("INSERT INTO \"backups\" (\"created_at\",\"updated_at\",\"deleted_at\",\"binding_id\",\"user_id\",\"name\",\"credentials\") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING \"id\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), bid, sql.NullString{String: testUserId, Valid: true}, name, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"created_at", "updated_at", "deleted_at", "binding_id", "user_id", "name", "credentials",
		}))
	dbPatcher.ExpectCommit()
	e.On("GetDB").Return(db)
	e.On("GetRandomId").Return(bid)
	link, err := GenerateBackupLink(ctx, e)

	return link, err
}

func TestGenerateBackupLink(t *testing.T) {
	testBid := "ttteesstt"
	testName := "testBackup"
	testUserId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"

	link, err := testGenerateBackupLink(common.ModeUpload, testBid, testName)
	require.Nil(t, err)
	require.NotNil(t, link)
	require.Equal(t, fmt.Sprintf("https://0.0.0.0/api/accounts/credentials/backup/%s/%s", testUserId, testBid), link.(BackupLinkOutput).Path)
	link2, err := testGenerateBackupLink(common.ModeDownload, testBid, testName)
	require.Nil(t, err)
	require.NotNil(t, link2)
	require.Equal(t, fmt.Sprintf("https://0.0.0.0/api/accounts/credentials/backup/%s/%s", testUserId, testBid), link2.(BackupLinkOutput).Path)
}

func TestCreateBackupCredentials(t *testing.T) {
	testCreds := []byte("nsclkasdnclkndklancanclamnclknacnncancncn=")
	testUserId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"
	testBid := "ttteesstt"
	testName := "testBackup"
	tP := &test.TestProvider{}

	rec := httptest.NewRecorder()
	ctx, eng := gin.CreateTestContext(rec)

	link, _ := testGenerateBackupLink(common.ModeUpload, testBid, testName)
	path, _ := strings.CutPrefix(link.(BackupLinkOutput).Path, "https://0.0.0.0")

	req, _ := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(testCreds))

	ctx.Request = req
	e := test.EnvObjMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.
		ExpectQuery("SELECT * FROM \"backups\" WHERE binding_id=$1 AND \"backups\".\"deleted_at\" IS NULL ORDER BY \"backups\".\"id\" LIMIT 1").
		WithArgs(testBid).
		WillReturnRows(sqlmock.NewRows([]string{
			"created_at", "updated_at", "deleted_at", "binding_id", "user_id", "name", "credentials",
		}).AddRow(time.Now(), time.Now(), nil, testBid, sql.NullString{String: testUserId, Valid: true}, testName, nil),
		)
	dbPatcher.ExpectBegin()
	dbPatcher.
		ExpectExec("UPDATE \"backups\" SET \"credentials\"=$1,\"updated_at\"=$2 WHERE binding_id=$3 AND \"backups\".\"deleted_at\" IS NULL").
		WithArgs(ValidBackupPayload{}, sqlmock.AnyArg(), testBid).
		WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectCommit()
	e.On("GetDB").Return(db)
	e.On("GetCryptoProvider").Return(tP)

	eng.PUT("api/accounts/credentials/backup/:id/:bid", common.ConstructResponse(CreateBackupCredentials, &e))
	eng.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	if rec.Code != http.StatusOK {
		tmp, _ := io.ReadAll(rec.Body)
		var res map[string]interface{}
		_ = json.Unmarshal(tmp, &res)
		assert.Equal(t, res, "success")
	}
}

func TestGetAllBackupCredentials(t *testing.T) {
	testCreds := []byte("nsclkasdnclkndklancanclamnclknacnncancncn=")
	testUserId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"
	testBid := "ttteesstt"
	testName := "testBackup"

	tP := &test.TestProvider{}
	keyId := types.CryptoIdentifier{
		KeyId: testUserId,
		CryptoContext: types.CryptoContext{
			Namespace: "account",
			Group:     "account",
			Context:   context.Background(),
		},
	}
	_ = tP.GenerateKey(types.CryptoKeyParameter{Identifier: keyId, KeyType: types.Aes256GCM})
	enCreds, _ := tP.Encrypt(keyId, testCreds)

	rec := httptest.NewRecorder()
	ctx, eng := gin.CreateTestContext(rec)
	req, _ := http.NewRequest(http.MethodGet, "/api/accounts/backup/all", nil)
	user := common.UserInfo{UserInfo: &gocloak.UserInfo{Sub: &testUserId}}
	req = req.WithContext(context.WithValue(req.Context(), common.UserKey, &user))
	ctx.Request = req
	e := test.EnvObjMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	jesusIsBorn, _ := time.Parse(time.UnixDate, "")
	dbPatcher.
		ExpectQuery("SELECT * FROM \"backups\" WHERE (user_id=$1 AND created_at >= $2) AND \"backups\".\"deleted_at\" IS NULL ORDER BY created_at DESC").
		WithArgs(testUserId, jesusIsBorn).
		WillReturnRows(sqlmock.
			NewRows(
				[]string{
					"created_at", "updated_at", "deleted_at", "binding_id", "user_id", "name", "credentials",
				}).
			AddRow(time.Now(), time.Now(), nil, testBid, sql.NullString{String: testUserId, Valid: true}, testName, enCreds),
		)
	e.On("GetDB").Return(db)
	e.On("GetCryptoProvider").Return(tP)

	eng.GET("/api/accounts/backup/all", common.ConstructResponse(GetAllBackupCredentials, &e))
	eng.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	tmp, _ := io.ReadAll(rec.Body)
	var res GetBackupsOutput
	_ = json.Unmarshal(tmp, &res)

	require.Equal(t, 1, len(res.Backups))
	require.Equal(t, testBid, res.Backups[0].BindingId)
	require.Equal(t, testUserId, res.Backups[0].UserId)
	require.Equal(t, testCreds, res.Backups[0].Credentials)
}

func TestGetBackupCredentials(t *testing.T) {
	testCreds := []byte("nsclkasdnclkndklancanclamnclknacnncancncn=")
	testUserId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"
	testBid := "ttteesstt"
	testName := "testBackup"

	tP := &test.TestProvider{}
	keyId := types.CryptoIdentifier{
		KeyId: testUserId,
		CryptoContext: types.CryptoContext{
			Namespace: "account",
			Group:     "account",
			Context:   context.Background(),
		},
	}
	_ = tP.GenerateKey(types.CryptoKeyParameter{Identifier: keyId, KeyType: types.Aes256GCM})
	enCreds, _ := tP.Encrypt(keyId, testCreds)

	rec := httptest.NewRecorder()
	ctx, eng := gin.CreateTestContext(rec)

	link, _ := testGenerateBackupLink(common.ModeUpload, testBid, testName)
	path, _ := strings.CutPrefix(link.(BackupLinkOutput).Path, "https://0.0.0.0")

	req, _ := http.NewRequest(http.MethodGet, path, nil)
	ctx.Request = req
	e := test.EnvObjMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.
		ExpectQuery("SELECT * FROM \"backups\" WHERE binding_id=$1 AND \"backups\".\"deleted_at\" IS NULL ORDER BY \"backups\".\"id\" LIMIT 1").
		WithArgs(testBid).
		WillReturnRows(sqlmock.
			NewRows(
				[]string{
					"created_at", "updated_at", "deleted_at", "binding_id", "user_id", "name", "credentials"}).
			AddRow(time.Now(), time.Now(), nil, sql.NullString{String: testBid, Valid: true}, testUserId, testName, enCreds),
		)
	dbPatcher.ExpectBegin()
	dbPatcher.
		ExpectQuery("INSERT INTO \"backups\" (\"created_at\",\"updated_at\",\"deleted_at\",\"binding_id\",\"user_id\",\"name\",\"credentials\") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING \"id\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sql.NullString{String: "", Valid: false}, testUserId, testName, enCreds).
		WillReturnRows(sqlmock.
			NewRows(
				[]string{
					"created_at", "updated_at", "deleted_at", "binding_id", "credentials"}).
			AddRow(time.Now(), time.Now(), nil, sql.NullString{String: "", Valid: false}, enCreds),
		)
	dbPatcher.ExpectCommit()
	e.On("GetDB").Return(db)
	e.On("GetCryptoProvider").Return(tP)

	eng.GET("api/accounts/credentials/backup/:id/:bid", common.ConstructResponse(GetBackupCredentials, &e))
	eng.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	tmp, _ := io.ReadAll(rec.Body)
	var res GetBackupsOutput
	_ = json.Unmarshal(tmp, &res)

	require.Equal(t, 1, len(res.Backups))
	require.Equal(t, testCreds, res.Backups[0].Credentials)
}

func TestDeleteBackupCredentials(t *testing.T) {
	testBid := "ttteesstt"
	rec := httptest.NewRecorder()
	ctx, eng := gin.CreateTestContext(rec)

	req, _ := http.NewRequest(http.MethodDelete, "/backup/delete/"+testBid, nil)
	ctx.Request = req
	e := test.EnvObjMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectBegin()
	dbPatcher.
		ExpectExec("UPDATE \"backups\" SET \"deleted_at\"=$1 WHERE binding_id=$2 AND \"backups\".\"deleted_at\" IS NULL").
		WithArgs(sqlmock.AnyArg(), testBid).
		WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectCommit()

	e.On("GetDB").Return(db)
	eng.DELETE("/backup/delete/:bid", common.ConstructResponse(DeleteBackup, &e))
	eng.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
