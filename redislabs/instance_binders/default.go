package instancebinders

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/pivotal-cf/brokerapi"
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

func (d *defaultBinder) Bind(instanceID string, bindingID string, persister persisters.StatePersister) (interface{}, error) {
	state, err := persister.Load()
	if err != nil {
		d.logger.Error("Failed to load the broker state", err)
		return nil, err
	}
	for _, instance := range state.AvailableInstances {
		if instance.ID == instanceID {
			creds := instance.Credentials
			return map[string]interface{}{
				"port":     creds.Port,
				"ip_list":  creds.IPList,
				"password": creds.Password,
			}, nil
		}
	}
	return nil, brokerapi.ErrInstanceDoesNotExist
}
