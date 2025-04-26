package config

import (
	"github.com/kelseyhightower/envconfig"
	cloud "gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/microservice/core/pkg/config"
	configPkg "gitlab.eclipse.org/eclipse/xfsc/libraries/microservice/core/pkg/config"
	"log"
	"time"
)

const (
	EnvVarPrefix = "ACCOUNT"
)

type TemplateConfiguration struct {
	config.BaseConfig `mapstructure:",squash"`
	Name              string `mapstructure:"serviceName" envconfig:"ACCOUNT_SERVICE_NAME"`

	Protocol cloud.ProtocolType `mapstructure:"protocol,omitempty" envconfig:"ACCOUNT_PROTOCOL"`

	Nats struct {
		WithNats     bool          `mapstructure:"withNats" envconfig:"ACCOUNT_NATS_WITHNATS"`
		Url          string        `mapstructure:"url" envconfig:"ACCOUNT_NATS_URL"`
		QueueGroup   string        `mapstructure:"queueGroup,omitempty" envconfig:"ACCOUNT_NATS_QUEUEGROUP"`
		TimeoutInSec time.Duration `mapstructure:"timeoutInSec,omitempty" envconfig:"ACCOUNT_NATS_TIMEOUTINSEC"`
	} `mapstructure:"nats"`

	Database struct {
		WithDB   bool   `mapstructure:"withDB,omitempty" envconfig:"ACCOUNT_DB_WITHDB" default:"false"`
		DBType   string `mapstructure:"dbType,omitempty" envconfig:"ACCOUNT_DB_TYPE" default:"postgres"`
		Host     string `mapstructure:"host,omitempty" envconfig:"ACCOUNT_DB_HOST"`
		Port     int    `mapstructure:"port,omitempty" envconfig:"ACCOUNT_DB_PORT"`
		User     string `mapstructure:"user,omitempty" envconfig:"ACCOUNT_DB_USER"`
		Password string `mapstructure:"password,omitempty" envconfig:"ACCOUNT_DB_PASSWORD"`
		DBName   string `mapstructure:"dbName,omitempty" envconfig:"ACCOUNT_DB_NAME"`
	} `mapstructure:"db,omitempty"`

	CloudEvents struct {
		Topics []string `mapstructure:"topics" envconfig:"ACCOUNT_CLOUDEVENTS_TOPICS"`
	} `mapstructure:"cloudevents,omitempty"`

	KeyCloak struct {
		Url              string        `mapstructure:"url" envconfig:"ACCOUNT_KEYCLOAK_URL"`
		Login            string        `mapstructure:"login" envconfig:"ACCOUNT_KEYCLOAK_LOGIN"`
		Password         string        `mapstructure:"password" envconfig:"ACCOUNT_KEYCLOAK_PASSWORD"`
		RealmName        string        `mapstructure:"realmName" envconfig:"ACCOUNT_KEYCLOAK_REALMNAME"`
		TokenTTL         time.Duration `mapstructure:"tokenTTL" envconfig:"ACCOUNT_KEYCLOAK_TOKENTTL"`
		ExcludeEndpoints string        `mapstructure:"excludeEndpoints" envconfig:"ACCOUNT_KEYCLOAK_EXCLUDEENDPOINTS"`
	} `mapstructure:"keycloak,omitempty"`

	Storage struct {
		Url      string `mapstructure:"url" envconfig:"ACCOUNT_STORAGE_URL"`
		KeyPath  string `mapstructure:"keyPath" envconfig:"ACCOUNT_STORAGE_KEYPATH"`
		WithAuth bool   `mapstructure:"withAuth" envconfig:"ACCOUNT_STORAGE_WITHAUTH"`
	} `mapstructure:"storage,omitempty"`

	BackupLinkTTL      time.Duration `mapstructure:"backupLinkTTL" envconfig:"ACCOUNT_BACKUPLINKTTL"`
	CredentialVerifier struct {
		Url string `mapstructure:"url" envconfig:"ACCOUNT_CREDENTIALVERIFIER_URL"`
	} `mapstructure:"credentialVerifier,omitempty"`

	DIDComm struct {
		Url string `mapstructure:"url" envconfig:"ACCOUNT_DIDCOMM_URL"`
	} `mapstructure:"didcomm,omitempty"`

	Signer struct {
		Url string `mapstructure:"url" envconfig:"ACCOUNT_SIGNER_URL"`
	} `mapstructure:"signer,omitempty"`

	CredentialRetrival struct {
		Url        string `mapstructure:"url" envconfig:"ACCOUNT_CREDENTIALRETRIEVAL_URL"`
		OfferTopic string `mapstructure:"offerTopic" envconfig:"ACCOUNT_CREDENTIALRETRIEVAL_OFFERTOPIC"`
	} `mapstructure:"credentialRetrieval,omitempty"`

	PluginDiscovery struct {
		Url string `mapstructure:"url" envconfig:"ACCOUNT_PLUGINDISCOVERY_URL"`
	} `mapstructure:"pluginDiscovery,omitempty"`
}

var ServerConfiguration TemplateConfiguration

func Init() {
	// Load Configuration
	err := configPkg.LoadConfig(EnvVarPrefix, &ServerConfiguration, nil)
	if err == nil {

		err = envconfig.Process(EnvVarPrefix, &ServerConfiguration)

		if err != nil {
			log.Fatalf("envconfig was not loaded: %t", err)
		}
	} else {
		log.Fatalf("config was not loaded: %t", err)
	}
}
