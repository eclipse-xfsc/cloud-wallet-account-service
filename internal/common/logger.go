package common

import (
	"gitlab.eclipse.org/eclipse/xfsc/libraries/microservice/core/pkg/logr"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"log"
	"sync"
)

var logger logr.Logger

var once sync.Once

func GetLogger() logr.Logger {
	once.Do(initLogger)
	return logger
}

func initLogger() {
	l, err := logr.New(config.ServerConfiguration.LogLevel, config.ServerConfiguration.IsDev, nil)
	if err != nil {
		log.Fatal(err)
	}
	logger = *l
}
