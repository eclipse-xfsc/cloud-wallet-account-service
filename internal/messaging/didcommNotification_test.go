package messaging

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/require"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
)

func TestHandleDIDCommNotification(t *testing.T) {
	t.Parallel()
	testUserId := "testId"
	remoteDid := "55566"

	record := DIDCommNotification{Account: testUserId, RemoteDID: remoteDid}
	data, _ := json.Marshal(record)
	e := event.New()
	e.SetID("testEventId")
	e.SetSource("test/didcommnotification")
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
	mock.ExpectQuery("INSERT INTO \"user_connections\" "+
		"(\"created_at\",\"updated_at\",\"deleted_at\",\"user_id\",\"remote_did\") "+
		"VALUES ($1,$2,$3,$4,$5) RETURNING \"id\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), testUserId, remoteDid).
		WillReturnRows(sqlmock.NewRows(
			[]string{
				"created_at", "updated_at", "deleted_at", "user_id", "remote_did"}))
	mock.ExpectCommit()
	err = HandleDIDCommNotification(e, envMock)
	require.Equal(t, nil, err)
}
