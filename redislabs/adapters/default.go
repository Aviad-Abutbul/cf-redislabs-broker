package adapters

import (
	"github.com/altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/altoros/cf-redislabs-broker/redislabs/persisters"
)

type DefaultCreator struct{}

func (d *DefaultCreator) Create(instanceID string, settings cluster.InstanceSettings, persister persisters.StatePersister) error {
	return nil
}

func (d *DefaultCreator) Destroy(instanceID string, persister persisters.StatePersister) error {
	return nil
}

func (d *DefaultCreator) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}
