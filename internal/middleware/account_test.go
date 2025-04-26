package middleware

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testToken = "testtokentesttoken"
var testSub = "testID"

func TestCheckExistenceAndGetUserData(t *testing.T) {
	response := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(response)

	c.Request, _ = http.NewRequest(http.MethodGet, "/deviceList", nil)
	c.Request.Header.Add("Authorization", testToken)

	mocked := &test.ProviderMock{}
	mockObj := mocked.On("GetUserInfo", context.Background(), testToken, "").
		Return(&gocloak.UserInfo{Sub: &testSub}, nil)
	defer func() { mockObj.Unset() }()
	engine.Use(CheckExistenceAndGetUserData(mocked))
	engine.GET("/deviceList", func(c *gin.Context) {
		c.JSON(http.StatusOK, "success")
	})
	engine.ServeHTTP(response, c.Request)
	require.Equal(t, http.StatusOK, response.Result().StatusCode)
}

func TestCreateCryptoKeysIfAccountIsNewError(t *testing.T) {
	response := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(response)

	c.Request, _ = http.NewRequest(http.MethodGet, "/deviceList", nil)
	envMock := &test.EnvObjMock{}
	engine.Use(CreateCryptoKeysIfAccountIsNew(envMock))
	engine.GET("/deviceList", func(c *gin.Context) {
		c.JSON(http.StatusOK, "success")
	})
	engine.ServeHTTP(response, c.Request)
	require.Equal(t, http.StatusForbidden, response.Result().StatusCode)
	envMock.AssertExpectations(t)
}

func TestCreateCryptoKeysIfAccountIsNew(t *testing.T) {
	response := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(response)

	c.Request, _ = http.NewRequest(http.MethodGet, "/deviceList", nil)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), common.UserKey, &common.UserInfo{UserInfo: &gocloak.UserInfo{Sub: &testSub}}))
	dbMock, _ := test.GetDBMock()
	envMock := &test.EnvObjMock{}
	envMock.
		On("GetDB").
		Return(dbMock).
		On("GetCryptoProvider").
		Return(&test.TestProvider{})

	engine.Use(CreateCryptoKeysIfAccountIsNew(envMock))
	engine.GET("/deviceList", func(c *gin.Context) {
		c.JSON(http.StatusOK, "success")
	})
	engine.ServeHTTP(response, c.Request)
	require.Equal(t, http.StatusOK, response.Result().StatusCode)
	envMock.AssertExpectations(t)
}
