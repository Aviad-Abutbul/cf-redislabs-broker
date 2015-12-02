package instancebinders

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
)

type default struct {
	conf config.Config
	logger lager.Logger
}

func NewDefault(conf config.Config, logger lager.Logger) *default {
	return &default{
		conf: conf,
		logger: logger,
	}
}

func (d *default) Unbind(instanceID string, bindingID string, persister persisters.StatePersister) error {
	return nil
}

func (d *default) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}

func (d *default) Bind(instanceID string, bindingID string, persister persisters.StatePersister) (redislabs.ServiceInstanceCredentials, error) {
	// load data from persister
	// WIP
	creds := redislabs.ServiceInstanceCredentials{
		Host:     "somo-address",
		Port:     8080,
		Password: "password",
	}
	return creds, nil
}
