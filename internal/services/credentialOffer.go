package services

import (
	"bytes"
	"encoding/json"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/credential"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"net/http"
	"strings"
	"sync"
	"time"
)

var credentialRetrieval *CredentialRetrieval

var onceCredentialRetrieval sync.Once

func GetCredentialRetrieval(client common.HttpClient) *CredentialRetrieval {
	onceCredentialRetrieval.Do(initCredentialRetrieval(client))
	return credentialRetrieval
}

func initCredentialRetrieval(client common.HttpClient) func() {
	return func() {
		credentialRetrieval = &CredentialRetrieval{url: config.ServerConfiguration.CredentialRetrival.Url, httpClient: client}
	}
}

type credentialRetrievalEndpoint string

const (
	getOffers    credentialRetrievalEndpoint = "/offering/list/:groupId"
	createOffer  credentialRetrievalEndpoint = "/offering/retrieve/:groupId"
	resolveOffer credentialRetrievalEndpoint = "/offering/clear/:groupId/:requestId"
)

type CredentialRetrieval struct {
	url        string
	httpClient common.HttpClient
}

func (c *CredentialRetrieval) GetOffers(groupId string) (*[]CredentialOffer, error) {
	var res []CredentialOffer
	path := strings.Replace(string(getOffers), ":groupId", groupId, 1)
	url := strings.Join([]string{c.url, path}, "")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return handleResponse(resp, &res)
}

func (c *CredentialRetrieval) CreateOffer(groupId string, offer CredentialOfferPayload) (string, error) {
	var res string
	path := strings.Replace(string(createOffer), ":groupId", groupId, 1)
	url := strings.Join([]string{c.url, path}, "")
	data, err := json.Marshal(offer)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	offerId, err := handleResponse(resp, &res)
	if err != nil {
		return "", err
	}
	return *offerId, nil
}

func (c *CredentialRetrieval) AcceptOffer(groupId string, requestId string, acceptanceData OfferAcceptanceData) (*AcceptedCredential, error) {
	var res AcceptedCredential
	path := strings.Replace(strings.Replace(string(resolveOffer), ":groupId", groupId, 1), ":requestId", requestId, 1)
	url := strings.Join([]string{c.url, path}, "")
	data, err := json.Marshal(acceptanceData)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return handleResponse(resp, &res)
}

type CredentialOfferPayload struct {
	CredentialOfferUri string `json:"credential_offer_uri,omitempty"`
	CredentialOffer    string `json:"credential_offer,omitempty"`
}

type CredentialOffer struct {
	GroupId   string                               `json:"groupId"`
	RequestId string                               `json:"requestId"`
	MetaData  credential.IssuerMetadata            `json:"metadata"`
	Offering  credential.CredentialOfferParameters `json:"offering"`
	Status    string                               `json:"status"`
	TimeStamp time.Time                            `json:"timestamp"`
}

type AcceptedCredential struct {
	Format          string      `json:"format"`
	Credential      interface{} `json:"credential,omitempty"`
	TransactionID   string      `json:"transaction_id,omitempty"`
	CNonce          string      `json:"c_nonce,omitempty"`
	CNonceExpiresIn int         `json:"c_nonce_expires_in,omitempty"`
}

type OfferAcceptanceData struct {
	Accept          bool   `json:"accept"`
	EncryptionKey   []byte `json:"encryptionKey,omitempty"`
	HolderKey       string `json:"holderKey"`
	HolderNamespace string `json:"holderNamespace"`
	HolderGroup     string `json:"holderGroup"`
}
