package redislabs

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/passwords"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/ldmberman/brokerapi"
	"github.com/pivotal-golang/lager"
)

type ServiceInstanceCreator interface {
	Create(instanceID string, settings cluster.InstanceSettings, persister persisters.StatePersister) error
	Update(instanceID string, params map[string]interface{}, persister persisters.StatePersister) error
	Destroy(instanceID string, persister persisters.StatePersister) error
	InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error)
}

type ServiceInstanceBinder interface {
	Bind(instanceID string, bindingID string, persister persisters.StatePersister) (interface{}, error)
	Unbind(instanceID string, bindingID string, persister persisters.StatePersister) error
	InstanceExists(instanceID string, persister persisters.StatePersister) (bool, error)
}

type serviceBroker struct {
	InstanceCreator ServiceInstanceCreator
	InstanceBinder  ServiceInstanceBinder
	StatePersister  persisters.StatePersister
	Config          config.Config
	Logger          lager.Logger
}

var (
	RedisPasswordLength = 48
)

func NewServiceBroker(
	instanceCreator ServiceInstanceCreator,
	instanceBinder ServiceInstanceBinder,
	statePersister persisters.StatePersister,
	conf config.Config,
	logger lager.Logger) *serviceBroker {

	return &serviceBroker{
		InstanceCreator: instanceCreator,
		InstanceBinder:  instanceBinder,
		StatePersister:  statePersister,
		Config:          conf,
		Logger:          logger,
	}
}

func (b *serviceBroker) Services() []brokerapi.Service {
	planList := []brokerapi.ServicePlan{}
	for _, p := range b.planDescriptions() {
		planList = append(planList, *p)
	}
	return []brokerapi.Service{
		brokerapi.Service{
			ID:            b.Config.ServiceBroker.ServiceID,
			Name:          b.Config.ServiceBroker.Name,
			Description:   b.Config.ServiceBroker.Description,
			Bindable:      true,
			Tags:          []string{"redislabs"},
			Plans:         planList,
			PlanUpdatable: true,
		},
	}
}

func (b *serviceBroker) Provision(instanceID string, provisionDetails brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	if provisionDetails.ID != b.Config.ServiceBroker.ServiceID {
		return false, ErrServiceDoesNotExist
	}
	settingsByID := b.instanceSettings()
	if _, ok := settingsByID[provisionDetails.PlanID]; !ok {
		return false, ErrPlanDoesNotExist
	}
	settings := settingsByID[provisionDetails.PlanID]
	password, err := passwords.Generate(RedisPasswordLength)
	if err != nil {
		b.Logger.Error("Failed to generate a password", err)
		return false, err
	}
	settings.Password = password
	return false, b.InstanceCreator.Create(instanceID, *settings, b.StatePersister)
}

func (b *serviceBroker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	if updateDetails.ID != b.Config.ServiceBroker.ServiceID {
		return false, ErrServiceDoesNotExist
	}

	settings := b.instanceSettings()
	params := map[string]interface{}{}

	if updateDetails.PlanID != "" {
		// If there is a request for a plan check whether it exists.
		plan, ok := settings[updateDetails.PlanID]
		if !ok {
			return brokerapi.IsAsync(false), ErrPlanDoesNotExist
		}
		// Record parameters coming from the plan change.
		params = map[string]interface{}{
			"memory_size":  plan.MemoryLimit,
			"replication":  plan.Replication,
			"shards_count": plan.ShardCount,
		}
	}

	// Check whether additional parameters are valid.
	if err := cluster.CheckUpdateParameters(updateDetails.Parameters); err != nil {
		b.Logger.Error("Invalid update JSON data", err)
		return brokerapi.IsAsync(false), err
	}
	// Record additional parameters.
	for param, value := range updateDetails.Parameters {
		params[param] = value
	}

	return brokerapi.IsAsync(false), b.InstanceCreator.Update(instanceID, params, b.StatePersister)
}

func (b *serviceBroker) Deprovision(instanceID string) error {
	return b.InstanceCreator.Destroy(instanceID, b.StatePersister)
}

func (b *serviceBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (interface{}, error) {
	return b.InstanceBinder.Bind(instanceID, bindingID, b.StatePersister)
}

// Redis Labs cluster does not support multitenancy within a single
// database. Therefore, the only goal of unbinding is to remove
// credentials from the application environment. Unbind exists as a part
// of the brokerapi.ServiceBroker interface and does not have to do
// any specific job in this context.
func (b *serviceBroker) Unbind(instanceID, bindingID string) error {
	return nil
}

func (b *serviceBroker) LastOperation(instanceID string) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

func (b *serviceBroker) planDescriptions() map[string]*brokerapi.ServicePlan {
	plansByID := map[string]*brokerapi.ServicePlan{}
	for _, plan := range b.Config.ServiceBroker.Plans {
		plansByID[plan.ID] = &brokerapi.ServicePlan{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
		}
	}
	return plansByID
}

func (b *serviceBroker) instanceSettings() map[string]*cluster.InstanceSettings {
	settingsByID := map[string]*cluster.InstanceSettings{}
	for _, plan := range b.Config.ServiceBroker.Plans {
		config := plan.ServiceInstanceConfig
		settingsByID[plan.ID] = &cluster.InstanceSettings{
			MemoryLimit: config.MemoryLimit,
			Replication: config.Replication,
			ShardCount:  config.ShardCount,
		}
	}
	return settingsByID
}
