package adapters

import (
	"github.com/altoros/redislabs-service-broker/redislabs/persisters"
)

type DefaultCreator struct {}

func (d *DefaultCreator) Create(instanceID string, persister persisters.StatePersister) error {
	return nil
}

func (d *DefaultCreator) Destroy(instanceID string, persister persisters.StatePersister) error {
	return nil
}

func (d *DefaultCreator) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}
