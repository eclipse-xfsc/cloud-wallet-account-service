package handlers

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListHistory(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req, _ := http.NewRequest(http.MethodGet, "/history", nil)
	testId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"
	user := common.UserInfo{UserInfo: &gocloak.UserInfo{Sub: &testId}}
	req = req.WithContext(context.WithValue(req.Context(), common.UserKey, &user))
	ctx.Request = req
	e := test.EnvObjMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectQuery("SELECT * FROM \"history_records\" WHERE user_id=$1 AND \"history_records\".\"deleted_at\" IS NULL").
		WithArgs(testId).
		WillReturnRows(sqlmock.
			NewRows(
				[]string{
					"created_at", "updated_at", "deleted_at", "user_id", "event_type", "message"}).
			AddRow(time.Now(), time.Now(), nil, testId, string(common.Issued), "Issue done"),
		)
	//populateDB(db)
	e.On("GetDB").Return(db)

	data, err := ListHistory(ctx, &e)
	require.Nil(t, err)
	res, ok := data.(ListHistoryOutput)
	if !ok {
		t.Error("unexpected output")
	}
	require.Equal(t, 1, len(res.Events))
}
