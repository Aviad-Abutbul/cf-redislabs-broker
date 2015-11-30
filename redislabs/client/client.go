package client

import (
	conf "github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/pivotal-golang/lager"
)

type RedislabsClient struct {
	config     *conf.RedislabsConfig
	httpClient *httpClient
	logger     lager.Logger
}

func NewRedislabsClient(config *conf.RedislabsConfig, logger lager.Logger) *RedislabsClient {
	logger.Info("Creating new redislabs client", lager.Data{
		"address": config.Address,
		"port":    config.Port,
	})
	httpClient := newHTTPClient(
		config.Auth.Username,
		config.Auth.Password,
		config.Address,
		config.Port,
		logger,
	)
	return &RedislabsClient{
		config:     config,
		logger:     logger,
		httpClient: httpClient,
	}
}
