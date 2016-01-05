package redislabs

import (
	"fmt"
	"strconv"

	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/passwords"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
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
		},
	}
}

func (b *serviceBroker) Provision(instanceID string, provisionDetails brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	if provisionDetails.ID != b.Config.ServiceBroker.ServiceID {
		return brokerapi.ProvisionedServiceSpec{IsAsync: false}, ErrServiceDoesNotExist
	}
	settingsByID := b.planSettings()
	if _, ok := settingsByID[provisionDetails.PlanID]; !ok {
		return brokerapi.ProvisionedServiceSpec{IsAsync: false}, ErrPlanDoesNotExist
	}
	planSettings := settingsByID[provisionDetails.PlanID]

	name, err := b.readDatabaseName(instanceID, provisionDetails)
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
	for param, value := range provisionDetails.Parameters {
		if param != "name" {
			settings[param] = tryParseInt(value)
		}
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

	if updateDetails.PlanID != "" {
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
		params[param] = tryParseInt(value)
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

func (b *serviceBroker) readDatabaseName(instanceID string, details brokerapi.ProvisionDetails) (string, error) {
	if details.Parameters == nil {
		return "", ErrDatabaseNameIsRequired
	}
	name, ok := details.Parameters["name"]
	if !ok || name == "" {
		return "", ErrDatabaseNameIsRequired
	}
	n := fmt.Sprintf("%s-%s", name, instanceID)
	if len(n) > RedisDatabaseNameLength {
		n = n[:RedisDatabaseNameLength]
	}
	return n, nil
}

func tryParseInt(value interface{}) interface{} {
	if floatValue, isFloat := value.(float64); isFloat {
		floatString := strconv.FormatFloat(floatValue, 'f', -1, 64)
		if maybeInt, err := strconv.ParseInt(floatString, 10, 64); err == nil {
			return maybeInt
		}
	}
	return value
}
