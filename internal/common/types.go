package common

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"gorm.io/gorm"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type UserInfo struct {
	*gocloak.UserInfo
}

func (u *UserInfo) ID() string {
	sub := u.UserInfo.Sub
	return *sub
}

type Env interface {
	IsHealthy() bool
	GetDB() *gorm.DB
	GetBroker(topic string) *cloudeventprovider.CloudEventProviderClient
	GetCryptoProvider() types.CryptoProvider
	GetNamespace() string
	AddBrokerSubscription(topic string, handler func(e event.Event))
	GetRandomId() string
	GetHttpClient() HttpClient
	AddBrokerPublication(topic string, e event.Event) error
	// SetSwaggerBasePath sets the base path that will be used by swagger ui for requests url generation
	SetSwaggerBasePath(path string)
	// SwaggerOptions swagger config options. See https://github.com/swaggo/gin-swagger?tab=readme-ov-file#configuration
	SwaggerOptions() []func(config *ginSwagger.Config)
}

type EndpointHandler func(*gin.Context, Env) (any, error)

type EventHandler func(event.Event, Env) error

type HistoryRecorder func(EndpointHandler, *gin.Context, Env, RecordEventType, string) (any, error)

type RecordEventType string

type RecordedFunc func(...any) error
