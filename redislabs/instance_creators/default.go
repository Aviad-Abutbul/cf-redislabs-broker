package instancecreators

import (
	"fmt"
	"sync"

	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/pivotal-golang/lager"
)

type defaultCreator struct {
	lock   sync.Mutex
	logger lager.Logger
	config config.Config
}

func NewDefault(config config.Config, logger lager.Logger) *defaultCreator {
	return &defaultCreator{
		config: config,
		logger: logger,
	}
}

func (d *defaultCreator) Create(instanceID string, settings cluster.InstanceSettings, persister persisters.StatePersister) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	// Load the broker state.
	d.logger.Info("Loading the broker state", lager.Data{
		"instance-id": instanceID,
	})
	state, err := persister.Load()
	if err != nil {
		d.logger.Fatal("Failed to load the broker state", err)
		return ErrFailedToLoadState
	}

	// Check whether the instance already exists.
	for _, s := range (*state).AvailableInstances {
		if s.ID == instanceID {
			d.logger.Error(fmt.Sprintf("Received a request to create an instance with ID %s that already exists", instanceID), ErrInstanceExists)
			return ErrInstanceExists
		}
	}

	// Ask the cluster to create a database.
	d.logger.Info("Creating a database", lager.Data{
		"instance-id": instanceID,
	})
	if err = d.createDatabase(); err != nil {
		return ErrFailedToCreateDatabase
	}

	s := persisters.ServiceInstance{
		ID: instanceID,
	}
	(*state).AvailableInstances = append((*state).AvailableInstances, s)
	// Save the new state.
	d.logger.Info("Saving the broker state", lager.Data{
		"instance-id": instanceID,
	})
	if err = persister.Save(state); err != nil {
		d.logger.Error("Failed to save the new state", err)
		return ErrFailedToSaveState
	}
	return nil
}

func (d *defaultCreator) Destroy(instanceID string, persister persisters.StatePersister) error {
	return nil
}

func (d *defaultCreator) InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error) {
	return false, nil
}

func (d *defaultCreator) createDatabase() error {
	return nil
}
