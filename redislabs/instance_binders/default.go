package instancebinders

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/pivotal-golang/lager"
)

type defaultBinder struct {
	conf   config.Config
	logger lager.Logger
}

func NewDefault(conf config.Config, logger lager.Logger) *defaultBinder {
	return &defaultBinder{
		conf:   conf,
		logger: logger,
	}
}

func (d *defaultBinder) Unbind(instanceID string, bindingID string, persister persisters.StatePersister) error {
	return nil
}

func (d *defaultBinder) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}

func (d *defaultBinder) Bind(instanceID string, bindingID string, persister persisters.StatePersister) (redislabs.ServiceInstanceCredentials, error) {
	// load data from persister
	// WIP
	creds := redislabs.ServiceInstanceCredentials{
		Host:     "somo-address",
		Port:     8080,
		Password: "password",
	}
	return creds, nil
}
