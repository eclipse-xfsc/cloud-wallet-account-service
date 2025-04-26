package services

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"io"
	"net/http"
)

type NilType struct {
	isNil bool
}

func handleResponse[T any](response *http.Response, dataAs *T) (*T, error) {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		if _, ok := any(dataAs).(*NilType); ok {
			return nil, nil
		}
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, dataAs)
		if err != nil {
			return nil, err
		}
		return dataAs, nil
	} else {
		var errBody = "<null>"
		if body, err := io.ReadAll(response.Body); err == nil && len(body) > 0 {
			errBody = string(body)
		}
		return nil, &common.ErrorResp{
			Err:          fmt.Errorf("not successful response received from external service. url: %s. received status: %s, received data: %s", response.Request.URL.String(), response.Status, errBody),
			Code:         http.StatusFailedDependency,
			ReceivedCode: response.StatusCode,
		}
	}
}
