package messaging

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging"
	msgCommon "gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/libraries/messaging/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"io"
	"net/http"
	"testing"
)

func TestHandlePresentationRequest(t *testing.T) {
	t.Parallel()
	testUserId := "testId"
	testMsg := "test is done"
	testRequestId := "testRequestId"
	presRequest, _ := json.Marshal(services.PresentationRequest{RequestId: testRequestId,
		PresentationDefinition: presentation.PresentationDefinition{}})
	creds, _ := json.Marshal(services.GetListCredentialModel{Groups: []presentation.FilterResult(
		[]presentation.FilterResult{presentation.FilterResult{
			Description: presentation.Description{Id: "", Name: "", Purpose: "", FormatType: ""},
			Credentials: map[string]interface{}{"cred1": "cred1"}}})})
	record := messaging.HistoryRecord{Reply: msgCommon.Reply{
		TenantId:  "",
		RequestId: testRequestId,
		Error:     nil,
	}, Message: testMsg, UserId: testUserId}
	data, _ := json.Marshal(record)
	e := event.New()
	e.SetID("testEventId")
	e.SetSource("test/history")
	e.SetType(string(common.PresentationRequest))
	err := e.SetData(ce.ApplicationJSON, data)
	if err != nil {
		t.Error("cannot set test event data")
	}

	mockDb, patch := test.GetDBMock()
	envMock := &test.EnvObjMock{}
	htt := test.HttpMock{}
	htt.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(presRequest))}).Once()
	htt.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(creds))}).Once()
	envMock.On("GetHttpClient").Return(&htt)
	envMock.On("GetDB").Return(mockDb)
	patch.ExpectExec("SET search_path TO accounts").
		WithoutArgs().
		WillReturnResult(sqlmock.NewResult(1, 1))
	patch.ExpectBegin()
	patch.ExpectQuery("INSERT INTO \"presentation_requests\" "+
		"(\"created_at\",\"updated_at\",\"deleted_at\",\"user_id\",\"request_id\") "+
		"VALUES ($1,$2,$3,$4,$5) RETURNING \"id\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), testUserId, testRequestId).
		WillReturnRows(sqlmock.NewRows(
			[]string{
				"created_at", "updated_at", "deleted_at", "user_id", "request_id"}))
	patch.ExpectCommit()
	err = HandlePresentationRequest(e, envMock)
	assert.Nil(t, err)
}
