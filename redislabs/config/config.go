package config

import (
	"os"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

type Config struct {
	Cluster       ClusterConfig       `yaml:"cluster"`
	ServiceBroker ServiceBrokerConfig `yaml:"broker"`
}

type ClusterConfig struct {
	Auth    AuthConfig `yaml:"auth"`
	Address string     `yaml:"address"`
}

type ServiceBrokerConfig struct {
	Auth        AuthConfig          `yaml:"auth"`
	Plans       []ServicePlanConfig `yaml:"plans"`
	ServiceID   string              `yaml:"service_id"`
	Port        int                 `yaml:"port"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Metadata    ServiceMetadata     `yaml:"metadata"`
}

type AuthConfig struct {
	Password string `yaml:"password"`
	Username string `yaml:"username"`
}

type ServicePlanConfig struct {
	ID                    string                `yaml:"id"`
	Name                  string                `yaml:"name"`
	Description           string                `yaml:"description"`
	Metadata              ServicePlanMetadata   `yaml:"metadata"`
	ServiceInstanceConfig ServiceInstanceConfig `yaml:"settings"`
}

type ServicePlanMetadata struct {
	Bullets []string `yaml:"bullets"`
}

type ServiceInstanceConfig struct {
	MemoryLimit int64    `yaml:"memory"`
	Replication bool     `yaml:"replication"`
	ShardCount  int64    `yaml:"shard_count"`
	Persistence string   `yaml:"persistence"`
	Snapshot    Snapshot `yaml:"snapshot"`
}

type Snapshot struct {
	Writes int `yaml:"writes"`
	Secs   int `yaml:"secs"`
}

type ServiceMetadata struct {
	Image               string `yaml:"image"`
	ProviderDisplayName string `yaml:"provider_display_name"`
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
