package requests

import (
	"bou.ke/monkey"
	"crypto/tls"
	"github.com/modern-go/reflect2"
	"net/http"
	"os"
	"testing"
)

func TestHttpClient(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch(os.ReadFile, func(file string) ([]byte, error) {
		return []byte{byte(1)}, nil
	})
	monkey.Patch(tls.LoadX509KeyPair, func(f string, k string) (tls.Certificate, error) {
		return tls.Certificate{}, nil
	})

	client, _ := HttpClient(false, http.Client{Timeout: 1000})
	actual := client.Timeout
	expected := 1000
	if int(actual) != expected {
		t.Errorf("actual %q, expected %q", actual, expected)
	}
	client, _ = HttpClient(true, http.Client{})
	actualT := client.Transport
	if reflect2.IsNil(actualT) {
		t.Errorf("actual nil, expected Transport with TLS config")
	}
}
