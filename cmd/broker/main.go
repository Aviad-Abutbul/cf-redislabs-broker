package main

import (
	"flag"
	"fmt"
	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	// "github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"
	"net/http"
	"os"
)

var (
	brokerConfigPath string
	brokerPidFile    string
)

func init() {
	flag.StringVar(&brokerConfigPath, "c", "", "Configuration File")

	flag.Parse()
}

func main() {

	brokerLogger := lager.NewLogger("redislabs-service-broker")
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	if brokerConfigPath == "" {
		brokerLogger.Fatal("No config file specified", nil)
	}

	brokerLogger.Info("Config File: " + brokerConfigPath)
	config, err := config.LoadFromFile(brokerConfigPath)

	if err != nil {
		brokerLogger.Fatal("Loading config file", err, lager.Data{
			"broker-config-path": brokerConfigPath,
		})
	}

	// instanceCreator := redislabs.ServiceInstanceCreator{}
	// instanceBinder := redislabs.ServiceInstanceBinder{}
	// persister := persisters.StatePersister{}

	serviceBroker := &redislabs.ServiceBroker{
		Logger: brokerLogger,
		// InstanceCreator: instanceCreator,
		// InstanceBinder:  instanceBinder,
		// StatePersister:  persister,
	}

	credentials := brokerapi.BrokerCredentials{
		Username: config.ServiceBroker.Auth.Username,
		Password: config.ServiceBroker.Auth.Password,
	}

	brokerAPI := brokerapi.New(serviceBroker, brokerLogger, credentials)
	http.Handle("/", brokerAPI)
	http.ListenAndServe(fmt.Sprintf(":%s", config.ServiceBroker.Port), nil)
}
