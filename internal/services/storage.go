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

	"github.com/ReneKroon/ttlcache"
	jwtUtil "github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/jwt"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/ssi/oid4vip/model/presentation"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/requests"
)

var storage *Storage

var onceStorage sync.Once

func GetStorage(client common.HttpClient) *Storage {
	onceStorage.Do(initStorage(client))
	return storage
}

func initStorage(client common.HttpClient) func() {
	return func() {
		cache := ttlcache.NewCache()
		storage = &Storage{userNonce: cache, url: config.ServerConfiguration.Storage.Url, httpClient: client}
	}
}

type Storage struct {
	url        string
	userNonce  *ttlcache.Cache
	httpClient common.HttpClient
	withAuth   bool
}

func (s *Storage) Register(auth string, accountId string) error {
	url := s.getEndpointUrl(accountId, register, "")
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", auth)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("register failed. %s", resp.Status)
	}
}

func (s *Storage) GetCredentials(auth string, accountId string, constraints *presentation.PresentationDefinition) ([]presentation.FilterResult, error) {
	return s.getCredentialsOrPresentation(credentials, auth, accountId, constraints)
}

func (s *Storage) GetPresentations(auth string, accountId string, constraints *presentation.PresentationDefinition) ([]presentation.FilterResult, error) {
	return s.getCredentialsOrPresentation(presentations, auth, accountId, constraints)
}

func (s *Storage) getCredentialsOrPresentation(urlPostfix storageEndpoint, auth string, accountId string, constraints *presentation.PresentationDefinition) ([]presentation.FilterResult, error) {

	url := s.getEndpointUrl(accountId, urlPostfix, "")
	logger.Info(fmt.Sprintf("request to storage service with %s", url))
	req, err := s.buildGetCredentialsReq(url, constraints)
	if err != nil {
		return nil, err
	}

	if s.withAuth {
		s = s.withSession(auth, accountId)
		nonce, _ := s.userNonce.Get(accountId)
		auth = authTokenWithNonce(auth, nonce.(string))
		req.Header.Add("Authorization", auth)
		req.Header.Add("Content-Type", "application/jose")
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var creds GetListCredentialModel
		err = json.Unmarshal(data, &creds)
		if err != nil {
			return nil, err
		}
		if s.withAuth {
			nonce, err := getNonce([]byte(creds.Receipt), auth)
			if err != nil {
				return nil, err
			}
			s.updateNonce(accountId, nonce)
		}

		return creds.Groups, nil

	} else {
		return nil, fmt.Errorf("get credentials failed. %s", resp.Status)
	}
}

type storageEndpoint string

const (
	register      storageEndpoint = "/device/registration/register"
	session       storageEndpoint = "/device/remote/session"
	credentials   storageEndpoint = "/credentials"
	presentations storageEndpoint = "/presentations"
)

func (s *Storage) buildGetCredentialsReq(urlStr string, constraints *presentation.PresentationDefinition) (req *http.Request, err error) {
	if constraints == nil {
		req, err = http.NewRequest(http.MethodPost, urlStr, nil)
		if err != nil {
			return nil, err
		}
	} else {
		jsonData, err := json.Marshal(constraints)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(http.MethodPost, urlStr, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
	}
	return req, err
}

func (s *Storage) getEndpointUrl(accountId string, endpoint storageEndpoint, id string) string {
	url := fmt.Sprintf("%s/%s%v", s.url, accountId, endpoint)
	if id == "" {
		return url
	}
	return fmt.Sprintf("%s/%s", url, id)
}

func (s *Storage) withSession(auth string, accountId string) *Storage {
	if _, ok := s.userNonce.Get(accountId); !ok {
		nonce, err := s.createSession(auth, accountId)
		if err != nil {
			s.updateNonce(accountId, nonce)
		}
	}
	return s
}

func (s *Storage) createSession(auth string, accountId string) (*TransactionModel, error) {
	url := s.getEndpointUrl(accountId, session, "")
	client, err := requests.HttpClient(false, *http.DefaultClient)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		transaction, err := getNonce(data, auth)
		if err != nil {
			return nil, err
		}
		return transaction, nil

	} else {
		return nil, fmt.Errorf("start session failed. %s", resp.Status)
	}
}

func (s *Storage) updateNonce(accountId string, transaction *TransactionModel) {
	exp := time.Unix(transaction.Expire, 0)
	s.userNonce.SetWithTTL(accountId, transaction.Nonce, exp.Sub(time.Now()))
}

type GetListCredentialModel struct {
	Groups  []presentation.FilterResult `json:"groups"`
	Receipt string                      `json:"receipt"`
}

type GetCredentialModel struct {
	Credentials map[string]interface{} `json:"credentials"`
	Receipt     string                 `json:"receipt"`
}

type TransactionModel struct {
	Nonce  string `json:"nonce"`
	Expire int64  `json:"expire"`
}

func getKey(jwtKey string) (jwk.Key, error) {
	t, err := jwtUtil.Parse(jwtKey, func(token *jwtUtil.Token) (interface{}, error) {
		return getPrivateKey()
	}, jwtUtil.WithValidMethods([]string{jwa.ES256.String()}))
	if err != nil {
		return nil, err
	}
	k := t.Header["jwk"].([]byte)
	return jwk.ParseKey(k)
}

func getPrivateKey() ([]byte, error) {
	data, err := os.ReadFile(config.ServerConfiguration.Storage.KeyPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func authTokenWithNonce(token string, nonce string) string {
	// todo implement logic that parses token, updates its claims with provided nonce and returns updated token as string
	return token
}

func getNonce(jweMessage []byte, jwtKey string) (*TransactionModel, error) {
	msg, err := jwe.Parse(jweMessage)
	if err != nil {
		return nil, err
	}
	key, err := getKey(jwtKey)
	if err != nil {
		return nil, err
	}
	data, err := jwt.DecryptJweMessage(msg, jwe.WithKey(jwa.ECDH_ES_A256KW, key))
	if err != nil {
		return nil, err
	}
	var trans TransactionModel
	err = json.Unmarshal(data, &trans)
	if err != nil {
		return nil, err
	} else {
		return &trans, nil
	}
}
