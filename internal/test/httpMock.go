package test

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type HttpMock struct {
	mock.Mock
}

func (m *HttpMock) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), nil
}
