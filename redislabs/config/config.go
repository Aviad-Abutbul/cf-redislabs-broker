package config

import (
	"os"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

type Config struct {
	Redislabs     RedislabsConfig     `yaml:"redislabs"`
	ServiceBroker ServiceBrokerConfig `yaml:"broker"`
}

type RedislabsConfig struct {
	Auth    AuthConfig `yaml:"auth"`
	Address string     `yaml:"address"`
	Port    int        `yaml:"port"`
}

type ServiceBrokerConfig struct {
	Auth      AuthConfig          `yaml:"auth"`
	Plans     []ServicePlanConfig `yaml:"plans"`
	ServiceID string              `yaml:"service_id"`
	Port      int                 `yaml:"port"`
	Name      string              `yaml:"name"`
}

type AuthConfig struct {
	Password string `yaml:"password"`
	Username string `yaml:"username"`
}

type ServicePlanConfig struct {
	ID                    string                `yaml:"id"`
	Name                  string                `yaml:"name"`
	Description           string                `yaml:"description"`
	ServiceInstanceConfig ServiceInstanceConfig `yaml:"settings"`
}

type ServiceInstanceConfig struct {
	MemoryLimit int64 `yaml:"memory"`
	Replication bool  `yaml:"replication"`
	ShardCount  int64 `yaml:"shard_count"`
}

func LoadFromFile(path string) (Config, error) {
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
