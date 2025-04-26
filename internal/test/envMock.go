package test

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/mock"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/docs"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"gorm.io/gorm"
)

type EnvObjMock struct {
	mock.Mock
}

func (env *EnvObjMock) IsHealthy() bool {
	return true
}

func (env *EnvObjMock) GetDB() *gorm.DB {
	args := env.Called()
	return args.Get(0).(*gorm.DB)
}

func (env *EnvObjMock) GetBroker(topic string) *cloudeventprovider.CloudEventProviderClient {
	args := env.Called()
	return args.Get(0).(*cloudeventprovider.CloudEventProviderClient)
}

func (env *EnvObjMock) GetCryptoProvider() types.CryptoProvider {
	args := env.Called()
	return args.Get(0).(types.CryptoProvider)
}

func (env *EnvObjMock) GetNamespace() string {
	return ""
}

func (env *EnvObjMock) AddBrokerSubscription(topic string, handler func(e event.Event)) {}

func (env *EnvObjMock) AddBrokerPublication(topic string, e event.Event) error { return nil }

func (env *EnvObjMock) GetRandomId() string {
	args := env.Called()
	return args.String(0)
}

func (env *EnvObjMock) GetHttpClient() common.HttpClient {
	args := env.Called()
	return args.Get(0).(common.HttpClient)
}

// SetSwaggerBasePath sets the base path that will be used by swagger ui for requests url generation
func (e *EnvObjMock) SetSwaggerBasePath(path string) {
	docs.SwaggerInfo.BasePath = path + common.BasePath
}

// SwaggerOptions swagger config options. See https://github.com/swaggo/gin-swagger?tab=readme-ov-file#configuration
func (e *EnvObjMock) SwaggerOptions() []func(config *ginSwagger.Config) {
	return []func(config *ginSwagger.Config){
		ginSwagger.DefaultModelsExpandDepth(10),
	}
}
