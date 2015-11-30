package instancecreators

import (
	"errors"
	"sync"

	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
)

type Default struct {
	lock sync.Mutex
}

func (d *Default) Create(instanceID string, settings cluster.InstanceSettings, persister persisters.StatePersister) error {
	d.lock.Lock()
	defer d.lock.Unlock()

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

func (d *Default) Destroy(instanceID string, persister persisters.StatePersister) error {
	return nil
}

func (d *Default) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}
