package config

import (
	"os"

	"github.com/pangpanglabs/goutils/behaviorlog"
	configutil "github.com/pangpanglabs/goutils/config"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/jwtutil"
	"github.com/sirupsen/logrus"
)

var config C

func Init(appEnv string, configPath string, options ...func(*C)) C {
	config.AppEnv = appEnv

	if configPath != "" {
		configutil.SetConfigPath(configPath)

	}

	if err := configutil.Read(appEnv, &config); err != nil {
		logrus.WithError(err).Warn("Fail to load config file")
	}

	if s := os.Getenv("JWT_SECRET"); s != "" {
		config.JwtSecret = s
		jwtutil.SetJwtSecret(s)
	}

	for _, option := range options {
		option(&config)
	}

	switch config.AppEnv {
	case "production", "qa":
		behaviorlog.SetLogLevel(logrus.ErrorLevel)
	case "staging":
		behaviorlog.SetLogLevel(logrus.InfoLevel)
	default:
		behaviorlog.SetLogLevel(logrus.DebugLevel)
	}

	return config
}

func Config() C {
	return config
}

type C struct {
	Database struct {
		Driver     string
		Connection string
		Logger     struct {
			Kafka echomiddleware.KafkaConfig
		}
	}
	BehaviorLog struct {
		Kafka echomiddleware.KafkaConfig
	}
	EventMessageBroker struct {
		Kafka echomiddleware.KafkaConfig
	}
	Services struct {
		StockApiUrl           string
		CalculatorApiUrl      string
		BenefitApiUrl         string
		ProductApiUrl         string
		PlaceManagementApiUrl string
	}
	AppEnv      string
	JwtSecret   string
	ServiceName string
}
