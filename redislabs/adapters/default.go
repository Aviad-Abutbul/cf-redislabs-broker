package adapters

import (
	"errors"
	"sync"

	"github.com/altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/altoros/cf-redislabs-broker/redislabs/persisters"
)

type DefaultCreator struct {
	sync.Mutex
}

func (d *DefaultCreator) Create(instanceID string, settings cluster.InstanceSettings, persister persisters.StatePersister) error {
	d.Lock()
	defer d.Unlock()

	state, err := persister.Load()
	if err != nil {
		return errors.New("") // TODO
	}

	// check if instance exists

	// make an API request

	// update state

	if err = persister.Save(state); err != nil {
		return errors.New("") // TODO
	}
	return nil
}

func (d *DefaultCreator) Destroy(instanceID string, persister persisters.StatePersister) error {
	return nil
}

func (d *DefaultCreator) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}
