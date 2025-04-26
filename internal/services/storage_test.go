package services

import (
	"bytes"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStorage_GetCredentials(t *testing.T) {
	creds := `{"groups":[{"credentials": {"cred1": "cred1"}, "receipt": ""}]}`
	body := io.NopCloser(bytes.NewReader([]byte(creds)))
	mocked := test.HttpMock{}
	resp := http.Response{StatusCode: http.StatusOK, Body: body}
	mocked.On("Do", mock.Anything).Return(&resp, nil)
	stor := GetStorage(&mocked)
	data, err := stor.GetCredentials("", "testId", nil)
	require.Nil(t, err)
	expected := []presentation.FilterResult(
		[]presentation.FilterResult{presentation.FilterResult{
			Description: presentation.Description{Id: "", Name: "", Purpose: "", FormatType: ""},
			Credentials: map[string]interface{}{"cred1": "cred1"}}})
	require.Equal(t, expected, data)
}
