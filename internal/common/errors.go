package common

const (
	CryptoError        = "crypto engine failed"
	UserNotFound       = "user not found by OIDC provider"
	HistoryRecordError = "could not save history record"
)

type ErrorResp struct {
	Err          error
	Code         int
	ReceivedCode int
}

func (e *ErrorResp) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

type ServerErrorResponse struct {
	Error string `json:"error"`
}
