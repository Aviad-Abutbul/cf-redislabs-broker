package instancebinders

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
)

type Default struct {
}

func (d *Default) Unbind(instanceID string, bindingID string, persister persisters.StatePersister) error {
	return nil
}

func (d *Default) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}

func (d *Default) Bind(instanceID string, bindingID string, persister persisters.StatePersister) (redislabs.ServiceInstanceCredentials, error) {
	// load data from persister
	// WIP
	creds := redislabs.ServiceInstanceCredentials{
		Host:     "somo-address",
		Port:     8080,
		Password: "password",
	}
	return creds, nil
}
