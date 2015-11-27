package config

import (
	"github.com/cloudfoundry-incubator/candiedyaml"
	"os"
)

type Config struct {
	Redislabs     redislabsConfiguration `yaml:"redislabs"`
	ServiceBroker brokerConfiguration    `yaml:"broker"`
}

type redislabsConfiguration struct {
	Auth AuthConfig `yaml:"auth"`
}

type brokerConfiguration struct {
	Auth      AuthConfig          `yaml:"auth"`
	Plans     []servicePlanConfig `yaml:"plans"`
	ServiceID string              `yaml:"service_id"`
	Port      int                 `yaml:"port"`
}

type AuthConfig struct {
	Password string `yaml:"password"`
	Username string `yaml:"username"`
}

type servicePlanConfig struct {
	ID               string                `yaml:"id"`
	Name             string                `yaml:"name"`
	Description      string                `yaml:"description"`
	InstanceSettings ServiceInstanceConfig `yaml:"settings"`
}

type ServiceInstanceConfig struct {
	MemoryLimit int64 `yaml:"memory_limit"`
	Replication bool  `yaml:"replication"`
	ShardCount  int64 `yaml:"shard_count"`
}

func LoadConfigFromFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := candiedyaml.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}
	// TODO: add validations here
	return config, nil
}
