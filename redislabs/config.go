package redislabs

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/pivotal-cf/brokerapi"
)

type Config struct {
	ServiceID    string              `yaml:"service_id"`
	DefaultPlans []ServicePlanConfig `yaml:"plans"`
}

type ServicePlanConfig struct {
	ID               string `yaml:"id"`
	Name             string
	Description      string
	InstanceSettings InstanceSettingsConfig `yaml:"cluster_properties"`
}

type InstanceSettingsConfig struct {
	MemoryLimit int64 `yaml:"memory_limit"`
	Replication bool  `yaml:"cluster"`
	ShardCount  int64 `yaml:"shard_count"`
}

func LoadPlanDescription(p ServicePlanConfig) brokerapi.ServicePlan {
	return brokerapi.ServicePlan{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
	}
}

func LoadInstanceSettings(p InstanceSettingsConfig) cluster.InstanceSettings {
	return cluster.InstanceSettings{
		MemoryLimit: p.MemoryLimit,
		Replication: p.Replication,
		ShardCount:  p.ShardCount,
	}
}
