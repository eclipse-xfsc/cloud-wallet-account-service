package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/requests"
)

// RequestBody struct represents the JSON structure for the request body
type InvitationRequestBody struct {
	Protocol   string            `json:"protocol"`
	Topic      string            `json:"topic"`
	Group      string            `json:"group"`
	EventType  string            `json:"eventType"`
	Properties map[string]string `json:"properties"`
}

type DIDCommConnection struct {
	RemoteDid     string            `json:"remoteDid"`
	RoutingKey    string            `json:"routingKey"`
	Protocol      string            `json:"protocol"`
	Topic         string            `json:"topic"`
	EventType     string            `json:"eventType"`
	Properties    map[string]string `json:"properties"`
	RecipientDids []string          `json:"recipientDids"`
	Added         time.Time         `json:"added"`
	Group         string            `json:"group"`
}

var didComm *DIDComm

var onceDidComm sync.Once

func GetDIDComm(client common.HttpClient) *DIDComm {
	onceDidComm.Do(initDIDComm(client))
	return didComm
}

func initDIDComm(client common.HttpClient) func() {
	var err error
	if client == nil {
		client, err = requests.HttpClient(false, *http.DefaultClient)
		if err != nil {
			logger.Error(err, "failed initialising didcomm service client")
			os.Exit(1)
		}
	}
	return func() {
		didComm = NewDIDComm(config.ServerConfiguration.DIDComm.Url, client)
	}
}

type DIDComm struct {
	url        string
	httpClient common.HttpClient
}

func NewDIDComm(url string, httpClient common.HttpClient) *DIDComm {
	return &DIDComm{
		url:        url,
		httpClient: httpClient,
	}
}

func (s *DIDComm) GetInviteLink(reqBody InvitationRequestBody) (string, error) {
	inviteURL := fmt.Sprintf("%s/admin/invitation", s.url)

	requestBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, inviteURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return "", err
	}

	response, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *DIDComm) GetConnectionList(accountID string, search string) ([]DIDCommConnection, error) {
	connectionURL := fmt.Sprintf("%s/admin/connections?group=%s", s.url, accountID)
	if search != "" {
		connectionURL = fmt.Sprintf("%s&search=%s", connectionURL, search)
	}
	req, err := http.NewRequest(http.MethodGet, connectionURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var remoteControlList []DIDCommConnection
	err = json.Unmarshal(body, &remoteControlList)
	if err != nil {
		return nil, err
	}

	return remoteControlList, nil
}

func (s *DIDComm) DeleteConnection(did string) error {
	connectionURL := fmt.Sprintf("%s/admin/connections/%s", s.url, did)

	req, err := http.NewRequest(http.MethodDelete, connectionURL, nil)
	if err != nil {
		return err
	}

	response, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}

func (s *DIDComm) BlockConnection(did string) error {
	connectionURL := fmt.Sprintf("%s/admin/connections/block/%s", s.url, did)

	req, err := http.NewRequest(http.MethodPost, connectionURL, nil)
	if err != nil {
		return err
	}

	response, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}
