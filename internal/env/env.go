package env

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/docs"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/connection"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/database"
	"gorm.io/gorm"
	"net/http"
	"os"
)

const (
	AccountServiceNamespace = "accountSpace"
)

var logger = common.GetLogger()

var env *EnvObj

var cryptoProvider types.CryptoProvider

func Init() {
	initCrypto()
	initEnv()
}

type EnvObj struct {
	db                  *gorm.DB
	brokerSubscriptions map[string]*cloudeventprovider.CloudEventProviderClient
}

func (env *EnvObj) IsHealthy() bool {
	db, err := env.db.DB()
	if err == nil {
		err = db.Ping()
	}
	return err == nil
}

func (env *EnvObj) GetDB() *gorm.DB {
	if env.db == nil {
		logger.Error(nil, "database connection was not initialised")
		return &gorm.DB{}
	}
	return env.db
}

func (env *EnvObj) GetBroker(topic string) *cloudeventprovider.CloudEventProviderClient {
	return env.brokerSubscriptions[topic]
}

func (env *EnvObj) GetCryptoProvider() types.CryptoProvider {
	return cryptoProvider
}

func (env *EnvObj) GetNamespace() string {
	return AccountServiceNamespace
}
func (env *EnvObj) AddBrokerSubscription(topic string, handler func(e event.Event)) {
	if config.ServerConfiguration.Nats.WithNats {
		broker, subscribe, err := connection.CloudEventsConnectionSubscribe(topic, handler)
		if err != nil {
			logger.Error(err, "subscription failed", "topic", topic)
			return
		}
		env.brokerSubscriptions[topic] = broker
		go func() {
			er := subscribe()
			if er != nil {
				logger.Error(er, "subscription failed", "topic", topic)
			}
		}()
		logger.Info("initialised broker subscription", "topic", topic)
	}
}

func (env *EnvObj) AddBrokerPublication(topic string, e event.Event) error {
	if config.ServerConfiguration.Nats.WithNats {
		_, publish, err := connection.CloudEventsConnectionPublish(topic, e)
		if err != nil {
			logger.Error(err, "publication failed", "topic", topic)
			return err
		}
		go func() error {
			er := publish()
			if er != nil {
				logger.Error(er, "publication failed", "topic", topic)
			}
			return er
		}()
	}
	return nil
}

func (env *EnvObj) GetRandomId() string {
	return uuid.New().String()
}

func (env *EnvObj) GetHttpClient() common.HttpClient {
	return http.DefaultClient
}

// SetSwaggerBasePath sets the base path that will be used by swagger ui for requests url generation
func (env *EnvObj) SetSwaggerBasePath(path string) {
	docs.SwaggerInfo.BasePath = path + common.BasePath
}

// SwaggerOptions swagger config options. See https://github.com/swaggo/gin-swagger?tab=readme-ov-file#configuration
func (env *EnvObj) SwaggerOptions() []func(config *ginSwagger.Config) {
	return []func(config *ginSwagger.Config){
		ginSwagger.DefaultModelsExpandDepth(10),
	}
}
func GetEnv() common.Env {
	return env
}

func initEnv() {
	envir := DefaultEnv()
	if config.ServerConfiguration.Database.WithDB {
		db, err := database.NewDatabaseConnection(common.DatabaseType(config.ServerConfiguration.Database.DBType))
		if err != nil {
			logger.Error(err, "connection to database failed")
			os.Exit(1)
		} else {
			envir.db = db
		}
	}

	env = &envir
}

func initCrypto() {
	cryptoProvider = core.CryptoEngine()
}

func DefaultEnv() EnvObj {
	envir := EnvObj{
		db:                  nil,
		brokerSubscriptions: make(map[string]*cloudeventprovider.CloudEventProviderClient),
	}
	return envir
}
