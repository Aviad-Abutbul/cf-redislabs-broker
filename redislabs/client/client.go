package client

import (
	"encoding/json"
	"fmt"
	conf "github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	"io/ioutil"
	"net/http"
	"net/url"
)

type RedislabsClient struct {
	config     *conf.RedislabsConfig
	httpClient *httpClient
	logger     *lager.Logger
}

func NewRedislabsClient(config *conf.RedislabsConfig, logger lager.Logger) RedislabsClient {
	logger.Info("Creating new redislabs client", lager.Data{address: address, port: port})
	httpClient := newHTTPClient(config.Address, config.Port)
	return &RedislabsClient{
		config:     config,
		logger:     logger,
		httpClient: httpClient,
	}
}
