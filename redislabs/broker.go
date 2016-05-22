package redislabs

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/RedisLabs/cf-redislabs-broker/redislabs/config"
	"github.com/RedisLabs/cf-redislabs-broker/redislabs/passwords"
	"github.com/RedisLabs/cf-redislabs-broker/redislabs/persisters"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"
)

type ServiceInstanceCreator interface {
	Create(instanceID string, settings map[string]interface{}, persister persisters.StatePersister) error
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
	RedisPasswordLength     = 48
	RedisDatabaseNameLength = 63
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
	b.Logger.Info("Serving a catalog request")
	return []brokerapi.Service{
		brokerapi.Service{
			ID:            b.Config.ServiceBroker.ServiceID,
			Name:          b.Config.ServiceBroker.Name,
			Description:   b.Config.ServiceBroker.Description,
			Bindable:      true,
			Tags:          []string{"redislabs"},
			Plans:         planList,
			PlanUpdatable: true,
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName:         b.Config.ServiceBroker.Metadata.DisplayName,
				ImageUrl:            b.Config.ServiceBroker.Metadata.Image,
				ProviderDisplayName: b.Config.ServiceBroker.Metadata.ProviderDisplayName,
			},
		},
	}
}

func (b *serviceBroker) Provision(instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	if details.ServiceID != b.Config.ServiceBroker.ServiceID {
		return brokerapi.ProvisionedServiceSpec{IsAsync: false}, ErrServiceDoesNotExist
	}
	settingsByID := b.planSettings()
	if _, ok := settingsByID[details.PlanID]; !ok {
		return brokerapi.ProvisionedServiceSpec{IsAsync: false}, ErrPlanDoesNotExist
	}
	planSettings := settingsByID[details.PlanID]

	// Unmarhal raw parameters
	var provisionParameters map[string]interface{}
	if len(details.RawParameters) > 0 {
		err := json.Unmarshal(details.RawParameters, &provisionParameters)
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{IsAsync: false}, brokerapi.ErrRawParamsInvalid
		}
	}

	name, err := b.readDatabaseName(instanceID, provisionParameters)
	if err != nil {
		b.Logger.Error("No database name was set", err)
		return brokerapi.ProvisionedServiceSpec{IsAsync: false}, err
	}

	settings := map[string]interface{}{
		"name": name,
	}
	// Record values coming from the plan.
	for param, value := range planSettings {
		settings[param] = value
	}

	// Record additional values. The name is excluded since we have
	// set it already.
	for param, value := range provisionParameters {
		settings[param] = castValue(value)
	}

	if _, ok := settings["authentication_redis_pass"]; !ok {
		password, err := passwords.Generate(RedisPasswordLength)
		if err != nil {
			b.Logger.Error("Failed to generate a password", err)
			return brokerapi.ProvisionedServiceSpec{IsAsync: false}, err
		}
		settings["authentication_redis_pass"] = password
	}

	return brokerapi.ProvisionedServiceSpec{IsAsync: false}, b.InstanceCreator.Create(instanceID, settings, b.StatePersister)
}

func (b *serviceBroker) Update(instanceID string, updateDetails brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	if updateDetails.ServiceID != b.Config.ServiceBroker.ServiceID {
		return false, ErrServiceDoesNotExist
	}

	settings := b.planSettings()
	params := map[string]interface{}{}

	if updateDetails.PlanID != updateDetails.PreviousValues.PlanID {
		// If there is a request for a plan check whether it exists.
		plan, ok := settings[updateDetails.PlanID]
		if !ok {
			return brokerapi.IsAsync(false), ErrPlanDoesNotExist
		}
		// Record parameters coming from the plan change.
		for param, value := range plan {
			params[param] = value
		}
	}

	// Record additional parameters.
	for param, value := range updateDetails.Parameters {
		params[param] = castValue(value)
	}

	return brokerapi.IsAsync(false), b.InstanceCreator.Update(instanceID, params, b.StatePersister)
}

func (b *serviceBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	return false, b.InstanceCreator.Destroy(instanceID, b.StatePersister)
}

func (b *serviceBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	b.Logger.Info("Looking for the service credentials", lager.Data{
		"instance-id": instanceID,
		"binding-id":  bindingID,
		"details":     details,
	})
	creds, err := b.InstanceBinder.Bind(instanceID, bindingID, b.StatePersister)
	return brokerapi.Binding{Credentials: creds}, err
}

// Redis Labs cluster does not support multitenancy within a single
// database. Therefore, the only goal of unbinding is to remove
// credentials from the application environment. Unbind exists as a part
// of the brokerapi.ServiceBroker interface and does not have to do
// any specific job in this context.
func (b *serviceBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails) error {
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
			Metadata: &brokerapi.ServicePlanMetadata{
				Bullets: plan.Metadata.Bullets,
			},
		}
	}
	return plansByID
}

func (b *serviceBroker) planSettings() map[string]map[string]interface{} {
	settingsByID := map[string]map[string]interface{}{}
	for _, plan := range b.Config.ServiceBroker.Plans {
		config := plan.ServiceInstanceConfig
		settings := map[string]interface{}{
			"memory_size":        config.MemoryLimit,
			"replication":        config.Replication,
			"shards_count":       config.ShardCount,
			"sharding":           config.ShardCount > 1,
			"implicit_shard_key": config.ShardCount > 1,
			"data_persistence":   config.Persistence,
		}
		if config.ShardCount > 1 {
			settings["shard_key_regex"] = []map[string]string{
				{"regex": `.*\{(?<tag>.*)\}.*`},
				{"regex": `(?<tag>.*)`},
			}
		}
		if config.Persistence == "snapshot" {
			settings["snapshot_policy"] = []map[string]int{{
				"writes": config.Snapshot.Writes,
				"secs":   config.Snapshot.Secs,
			}}
		}
		settingsByID[plan.ID] = settings
	}
	return settingsByID
}

func (b *serviceBroker) readDatabaseName(instanceID string, params map[string]interface{}) (string, error) {
	var nameParam interface{}

	nameParam, ok := params["name"]
	if !ok || nameParam == nil {
		nameParam = "cf"
	}

	name := fmt.Sprintf("%s-%s", nameParam, instanceID)
	if len(name) > RedisDatabaseNameLength {
		name = name[:RedisDatabaseNameLength]
	}

	return name, nil
}

func castValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// try int
		intValue, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			return intValue
		}

		// try float
		floatValue, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return floatValue
		}

		// try bool
		boolVal, err := strconv.ParseBool(v)
		if err == nil {
			return boolVal
		}
	}
	return value
}
