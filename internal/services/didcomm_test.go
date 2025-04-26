package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
)

// mockHttpClientGenerator is a function that generates a mockHttpClient for testing.
func mockHttpClientGenerator(expectedPath string, responseStatusCode int, responseBody string) common.HttpClient {
	return &mockHttpClient{
		expectedPath:       expectedPath,
		responseStatusCode: responseStatusCode,
		responseBody:       responseBody,
	}
}

// mockHttpClient implements the HttpClient interface for testing.
type mockHttpClient struct {
	expectedPath       string
	responseStatusCode int
	responseBody       string
}

// Do is the implementation of the HttpClient interface for mockHttpClient.
func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Println(req.URL.String())
	if req.URL.Path == m.expectedPath {
		return &http.Response{
			StatusCode: m.responseStatusCode,
			Body:       io.NopCloser(bytes.NewBufferString(m.responseBody)),
		}, nil
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}, nil
}

func TestGetInviteLink(t *testing.T) {
	expectedLink := "https://didcommhost.com/invitelink"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedLink))
	}))
	defer ts.Close()

	expectedPath := "/admin/invitation"
	responseStatusCode := http.StatusOK

	mockClient := mockHttpClientGenerator(expectedPath, responseStatusCode, expectedLink)

	service := NewDIDComm(ts.URL, mockClient)

	service.url = ts.URL

	reqBody := InvitationRequestBody{
		Protocol:  "nats",
		Topic:     "remotecontrol.xxx",
		Group:     "354348545845845",
		EventType: "device.connection",
		Properties: map[string]string{
			"account":  "354348545845845",
			"greeting": "hello-world",
		},
	}

	link, err := service.GetInviteLink(reqBody)
	require.NoError(t, err)

	if link != expectedLink {
		t.Errorf("Expected link: %s, got: %s", expectedLink, link)
	}
}

func TestGetConnectionList_WithoutProperties(t *testing.T) {
	// Set up a test server to mock the external service's response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(listConnections))
	}))
	defer ts.Close()

	expectedPath := "/admin/connections"
	responseStatusCode := http.StatusOK

	mockClient := mockHttpClientGenerator(expectedPath, responseStatusCode, listConnections)

	service := NewDIDComm(ts.URL, mockClient)

	service.url = ts.URL

	list, err := service.GetConnectionList("354348545845845", "")
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}

func TestGetConnectionList_WithProperties(t *testing.T) {
	// Set up a test server to mock the external service's response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		searchParam := r.URL.Query().Get("search")
		fmt.Println(searchParam)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(listConnections))
	}))
	defer ts.Close()

	expectedPath := "/admin/connections"
	responseStatusCode := http.StatusOK

	mockClient := mockHttpClientGenerator(expectedPath, responseStatusCode, listConnections)

	service := NewDIDComm(ts.URL, mockClient)

	service.url = ts.URL

	list, err := service.GetConnectionList("354348545845845", "12345")
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}

const listConnections = `[
	{
		"remoteDid": "555",
		"routingKey": "",
		"protocol": "nats",
		"topic": "remotecontrol.xxx",
		"eventType": "test",
		"properties": {
			"account": "354348545845845",
			"greeting": "hello-world"
		},
		"recipientDids": [
			"did:peer:2.Ez6LSjCAQ65hi4awjmGTkAJvUpVUJheheWjqsvjMh27UgtC8A.Vz6Mkoww2QpMsjDat3RsnfpAKsJSQYDbxXc1UYCreKVRdJXBQ.SeyJ0IjoiZG0iLCJzIjp7InVyaSI6Imh0dHBzOi8vY2xvdWQtd2FsbGV0Lnhmc2MuZGV2L21lc3NhZ2UvcmVjZWl2ZSIsImEiOlsiZGlkY29tbS92MiJdLCJyIjpbXX19"
		],
		"added": "2024-03-12T12:09:00.679Z",
		"group": ""
	}
]`
