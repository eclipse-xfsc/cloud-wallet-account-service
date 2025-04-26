package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePresentation(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, eng := gin.CreateTestContext(rec)
	testRequestId := "tttteeeesssstttt"
	testSignKey := "testKey"
	testDidId := "did:test:123"
	responsePayload := CreateProofPayload{
		SignKeyId: testSignKey,
		Filters:   []presentation.FilterResult{},
	}
	responsePayloadBytes, _ := json.Marshal(responsePayload)
	didList := services.ListDidResponse{List: []services.ListDidItem{services.ListDidItem{
		Did:  testDidId,
		Name: testSignKey,
	}}}
	didListBytes, _ := json.Marshal(didList)
	did := services.DidDocument{
		ID:                 testDidId,
		Controller:         testDidId,
		VerificationMethod: []*services.DIDVerificationMethod{},
		Service:            []*services.ServiceEndpoint{},
	}
	didBytes, _ := json.Marshal(did)

	req, _ := http.NewRequest(http.MethodPost, "/proof/"+testRequestId, bytes.NewBuffer(responsePayloadBytes))
	testUserId := "4c216ab0-a91a-413f-8e97-a32eee7a4ef4"
	user := common.UserInfo{UserInfo: &gocloak.UserInfo{Sub: &testUserId}}
	req = req.WithContext(context.WithValue(req.Context(), common.UserKey, &user))
	ctx.Request = req
	e := test.EnvObjMock{}
	htt := test.HttpMock{}
	db, dbPatcher := test.GetDBMock()
	dbPatcher.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectBegin()
	dbPatcher.ExpectExec("UPDATE \"presentation_requests\" SET \"deleted_at\"=$1 WHERE (user_id=$2 AND request_id=$3) AND \"presentation_requests\".\"deleted_at\" IS NULL").
		WithArgs(sqlmock.AnyArg(), testUserId, testRequestId).
		WillReturnResult(sqlmock.NewResult(1, 1))
	dbPatcher.ExpectCommit()
	e.On("GetDB").Return(db)
	htt.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(didListBytes))}).Once()
	htt.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(didBytes))}).Once()
	htt.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}).Once()
	e.On("GetHttpClient").Return(&htt)
	eng.POST("/proof/:id", common.ConstructResponse(CreatePresentation, &e))
	eng.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
