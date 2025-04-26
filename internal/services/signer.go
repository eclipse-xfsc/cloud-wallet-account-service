package services

import (
	"errors"
	"fmt"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"net/http"
	"sync"
)

var signer *Signer

var onceSigner sync.Once

func GetSigner(client common.HttpClient) *Signer {
	onceSigner.Do(initSigner(client))
	return signer
}

func initSigner(client common.HttpClient) func() {
	return func() {
		signer = &Signer{url: config.ServerConfiguration.Signer.Url, httpClient: client}
	}
}

type signerEndpoint string

const (
	listDidDocs signerEndpoint = "/v1/did/list"
	getDidDoc   signerEndpoint = "/v1/did/document"
)

const (
	NamespaceHeader string = "x-namespace"
	GroupHeader     string = "x-group"
	DIDHeader       string = "x-did"
)

type Signer struct {
	url        string
	httpClient common.HttpClient
}

type ListDidItem struct {
	Did  string `json:"did"`
	Name string `json:"name"`
}

type ListDidResponse struct {
	List []ListDidItem `json:"list"`
}

// DidDocument is the result type of the signer service didDoc method.
type DidDocument struct {
	// did of the document
	ID string
	// controller of the document
	Controller string
	// methods of the document
	VerificationMethod []*DIDVerificationMethod
	// service endpoints
	Service []*ServiceEndpoint
}

type ServiceEndpoint struct {
	// did of the document
	ID string
	// type of endpoint
	Type string
	// Endpoint URL
	ServiceEndpoint string
}

// DIDVerificationMethod Public Key represented as DID Verification Method.
type DIDVerificationMethod struct {
	// ID of verification method.
	ID string
	// Type of verification method key.
	Type string
	// Controller of verification method specified as DID.
	Controller string
	// Public Key encoded in JWK format.
	PublicKeyJwk any
}

func (s *Signer) ListDidDocs(namespace string, group string) (*ListDidResponse, error) {
	url := fmt.Sprintf("%s%s", s.url, listDidDocs)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(NamespaceHeader, namespace)
	req.Header.Add(GroupHeader, group)
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var didList ListDidResponse
	resp, err := handleResponse(res, &didList)
	if err != nil {
		var errResp *common.ErrorResp
		if errors.As(err, &errResp) && errResp.ReceivedCode == http.StatusNotFound {
			return &ListDidResponse{List: make([]ListDidItem, 0)}, nil
		}
		return nil, errors.Join(errors.New("failed to get did list"), err)
	}
	return resp, nil
}

func (s *Signer) GetDidDoc(id string, namespace string, group string) (*DidDocument, error) {
	url := fmt.Sprintf("%s%s", s.url, getDidDoc)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(NamespaceHeader, namespace)
	req.Header.Add(GroupHeader, group)
	req.Header.Add(DIDHeader, id)
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var did DidDocument
	return handleResponse(res, &did)
}
