package main

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
	serverPkg "gitlab.eclipse.org/eclipse/xfsc/libraries/microservice/core/pkg/server"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/api"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/env"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/messaging"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/middleware"
)

var logger = common.GetLogger()

func init() {
	config.Init()
	env.Init()
}

// @title			Account service API
// @version		1.0
// @description	API Gateway for the personal credential manager cloud services
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:8080

func main() {
	envir := env.GetEnv()
	logger.Info("Starting account service", "mode", config.ServerConfiguration.ServerMode)
	server := serverPkg.New(envir, config.ServerConfiguration.ServerMode)
	// broker subscriptions are added here
	registerBrokerHandlers(envir)
	defer closeBrokers(envir)
	// routes and middleware are added here
	server.Add(func(group *gin.RouterGroup) {
		baseGroup := group.Group(common.BasePath)
		baseGroup.Use(middleware.CheckExistenceAndGetUserData())
		baseGroup.Use(middleware.CreateCryptoKeysIfAccountIsNew())
		api.DeviceRoutes(baseGroup, envir)
		api.HistoryRoutes(baseGroup, envir)
		api.KmsRoutes(baseGroup, envir)
		api.SettingsRoutes(baseGroup, envir)
		api.CredentialRoutes(baseGroup, envir)
		api.PresentationsRoutes(baseGroup, envir)
		api.ConfiguationsRoutes(baseGroup, envir)
		api.PluginRoutes(baseGroup, envir)
	})

	err := server.Run(config.ServerConfiguration.ListenPort)
	if err != nil {
		panic(err)
	}
}

func registerBrokerHandlers(envir common.Env) {
	envir.AddBrokerSubscription("didcomm-connector-invitation", func(e event.Event) {
		logger.Info("received event", "topic", "didcomm-connector-invitation", "id", e.ID(), "type", e.Type())
		messaging.HandleError(messaging.HandleDIDCommNotification, e, envir, func(err error) {
			logger.Error(err, "failure during create key event handling")
		})
	})
	envir.AddBrokerSubscription("accounts.record", func(e event.Event) {
		logger.Info("received event", "topic", "accounts.record", "id", e.ID(), "type", e.Type())
		messaging.HandleError(messaging.HandleHistoryRecord, e, envir, nil)
		messaging.HandleError(messaging.HandlePresentationRequest, e, envir, nil)
		messaging.HandleError(messaging.HandleCreateKey, e, envir, func(err error) {
			logger.Error(err, "failure during create key event handling")
		})
	})

}

func closeBrokers(envir common.Env) {
	for _, topic := range []string{"accounts.record", "didcomm-connector-invitation"} {
		envir.GetBroker(topic).Close()
	}
}
