package messaging

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/require"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"testing"
)

func TestHandleHistoryRecord(t *testing.T) {
	t.Parallel()
	testUserId := "testId"
	testMsg := "test is done"

	record := messaging.HistoryRecord{Message: testMsg, UserId: testUserId}
	data, _ := json.Marshal(record)
	e := event.New()
	e.SetID("testEventId")
	e.SetSource("test/history")
	e.SetType(string(common.Issued))
	err := e.SetData(ce.ApplicationJSON, data)
	if err != nil {
		t.Error("cannot set test event data")
	}

	mockDb, mock := test.GetDBMock()
	envMock := &test.EnvObjMock{}
	envMock.On("GetDB").Return(mockDb)
	mock.ExpectExec("SET search_path TO accounts").
		WithoutArgs().
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO \"history_records\" "+
		"(\"created_at\",\"updated_at\",\"deleted_at\",\"user_id\",\"event_type\",\"message\") "+
		"VALUES ($1,$2,$3,$4,$5,$6) RETURNING \"id\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), testUserId, string(common.Issued), testMsg).
		WillReturnRows(sqlmock.NewRows(
			[]string{
				"created_at", "updated_at", "deleted_at", "user_id", "event_type", "message"}))
	mock.ExpectCommit()
	err = HandleHistoryRecord(e, envMock)
	require.Equal(t, nil, err)
}
