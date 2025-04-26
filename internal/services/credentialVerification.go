package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"net/http"
	"sync"
	"time"
)

var credentialVerification *CredentialVerification

var onceCredentialVerification sync.Once

func GetCredentialVerification(client common.HttpClient) *CredentialVerification {
	onceCredentialVerification.Do(initCredentialVerification(client))
	return credentialVerification
}

func initCredentialVerification(client common.HttpClient) func() {
	return func() {
		credentialVerification = &CredentialVerification{url: config.ServerConfiguration.CredentialVerifier.Url, httpClient: client}
	}
}

type credentialVerificationEndpoint string

const (
	proofEndpoint        credentialVerificationEndpoint = "/proofs/proof/%s"
	proofRequestEndpoint credentialVerificationEndpoint = "/proofs/proof/request/%s"
	assignPresentation   credentialVerificationEndpoint = "/proofs/proof/%s/assign/%s"
)

type CredentialVerification struct {
	url        string
	httpClient common.HttpClient
}

type PresentationRequest struct {
	Region                 string                              `json:"region"`
	Country                string                              `json:"country"`
	Id                     string                              `json:"id"`
	RequestId              string                              `json:"requestId"`
	GroupId                string                              `json:"groupid"`
	PresentationDefinition presentation.PresentationDefinition `json:"presentationDefinition"`
	Presentation           []interface{}                       `json:"presentation"`
	RedirectUri            string                              `json:"redirectUri"`
	ResponseUri            string                              `json:"responseUri"`
	ResponseMode           string                              `json:"responseMode"`
	ResponseType           string                              `json:"responseType"`
	State                  string                              `json:"state"`
	LastUpdateTimeStamp    time.Time                           `json:"lastUpdateTimeStamp"`
	Nonce                  string                              `json:"nonce"`
	ClientId               string                              `json:"clientId"`
}

type Proof struct {
	Payload       []presentation.FilterResult
	SignNamespace string
	SignKey       string
	SignGroup     string
	HolderDid     string
}

func (cv *CredentialVerification) getEndpointUrl(endpoint credentialVerificationEndpoint, id ...any) string {
	route := string(endpoint)
	if len(id) > 0 {
		route = fmt.Sprintf(route, id...)
	}
	return fmt.Sprintf("%s%s", cv.url, route)
}

func (cv *CredentialVerification) GetProofRequest(requestId string) (*PresentationRequest, error) {
	req, err := http.NewRequest(http.MethodGet, cv.getEndpointUrl(proofEndpoint, requestId), nil)
	if err != nil {
		return nil, err
	}
	resp, err := cv.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var res PresentationRequest
	return handleResponse(resp, &res)
}

func (cv *CredentialVerification) GetProofRequestByProofRequestId(proofRequestId string) (*PresentationRequest, error) {
	req, err := http.NewRequest(http.MethodGet, cv.getEndpointUrl(proofRequestEndpoint, proofRequestId), nil)
	if err != nil {
		return nil, err
	}
	resp, err := cv.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var res PresentationRequest
	return handleResponse(resp, &res)
}

func (cv *CredentialVerification) CreateProof(requestId string, filterResults []presentation.FilterResult, namespace string, group string, signKey string, holderDid string) error {
	proof := Proof{
		Payload:       filterResults,
		SignNamespace: namespace,
		SignKey:       signKey,
		SignGroup:     group,
		HolderDid:     holderDid,
	}
	data, err := json.Marshal(proof)
	body := bytes.NewBuffer(data)
	//logger.Info("sending proof", "data", body.String())
	req, err := http.NewRequest(http.MethodPost, cv.getEndpointUrl(proofRequestEndpoint, requestId), body)
	if err != nil {
		return err
	}
	resp, err := cv.httpClient.Do(req)
	if err != nil {
		return err
	}
	var res = NilType{}
	_, err = handleResponse(resp, &res)
	return err
}

func (cv *CredentialVerification) AssignProof(requestId string, userId string) error {
	req, err := http.NewRequest(http.MethodPut, cv.getEndpointUrl(assignPresentation, requestId, userId), nil)
	if err != nil {
		return err
	}
	resp, err := cv.httpClient.Do(req)
	if err != nil {
		return err
	}

	var res = NilType{}
	_, err = handleResponse(resp, &res)
	return err
}
