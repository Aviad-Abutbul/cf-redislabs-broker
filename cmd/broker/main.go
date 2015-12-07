package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/instance_binders"
	"github.com/Altoros/cf-redislabs-broker/redislabs/instance_creators"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/ldmberman/brokerapi"
	"github.com/pivotal-golang/lager"
)

var (
	localPersisterPath = path.Join(os.Getenv("HOME"), ".redislabs-broker", "state.json")

	brokerConfigPath string
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
		brokerLogger.Error("No config file specified", nil)
		return
	}

	brokerLogger.Info("Config File: " + brokerConfigPath)

	conf, err := config.LoadFromFile(brokerConfigPath)
	if err != nil {
		brokerLogger.Fatal("Loading config file", err, lager.Data{
			"broker-config-path": brokerConfigPath,
		})
	}

	serviceBroker := redislabs.NewServiceBroker(
		instancecreators.NewDefault(conf, brokerLogger),
		instancebinders.NewDefault(conf, brokerLogger),
		persisters.NewLocalPersister(localPersisterPath),
		conf,
		brokerLogger,
	)

	credentials := brokerapi.BrokerCredentials{
		Username: conf.ServiceBroker.Auth.Username,
		Password: conf.ServiceBroker.Auth.Password,
	}

	brokerAPI := brokerapi.New(serviceBroker, brokerLogger, credentials)
	http.Handle("/", brokerAPI)
	http.ListenAndServe(fmt.Sprintf(":%s", conf.ServiceBroker.Port), nil)
}
