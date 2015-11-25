package redislabs

import "github.com/pivotal-cf/brokerapi"

type Config struct {
	DefaultPlans []ServicePlanConfig `yaml:"plans"`
}

type ServicePlanConfig struct {
	ID          string `yaml:"id"`
	Name        string
	Description string
}

func LoadServicePlan(p ServicePlanConfig) brokerapi.ServicePlan {
	return brokerapi.ServicePlan{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
	}
}
