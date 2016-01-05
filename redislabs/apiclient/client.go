package apiclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	"github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/httpclient"
	"github.com/pivotal-golang/lager"
)

type (
	apiClient struct {
		conf   config.Config
		logger lager.Logger
	}

	errorResponse struct {
		ErrorMessage string `json:"description"`
		ErrorCode    string `json:"error_code"`
	}

	statusResponse struct {
		UID        int      `json:"uid"`
		Password   string   `json:"authentication_redis_pass"`
		IPList     []string `json:"endpoint_ip"`
		DNSAddress string   `json:"dns_address_master"`
		Status     string   `json:"status"`
	}
)

var (
	DatabasePollingInterval = 500 // milliseconds
)

func New(conf config.Config, logger lager.Logger) *apiClient {
	return &apiClient{
		conf:   conf,
		logger: logger,
	}
}

func (c *apiClient) CreateDatabase(settings map[string]interface{}) (chan cluster.InstanceCredentials, error) {
	bytes, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	httpClient := c.httpClient()
	c.logger.Info("Sending a database creation request", lager.Data{
		"settings": settings,
	})
	res, err := httpClient.Post("/v1/bdbs", httpclient.HTTPPayload(bytes))
	if err != nil {
		c.logger.Error("Failed to perform a database creation request", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		payload, err := c.parseErrorResponse(res)
		if err != nil {
			return nil, err
		}
		err = fmt.Errorf(payload.ErrorMessage)
		c.logger.Error("Failed to create a database", err)
		return nil, err
	}

	c.logger.Info("Database creation has been scheduled")
	ch := make(chan cluster.InstanceCredentials)
	go func() {
		var payload statusResponse
		for {
			time.Sleep(time.Duration(DatabasePollingInterval) * time.Millisecond)

			payload, err = c.parseStatusResponse(res)
			if err != nil {
				return
			}
			if payload.Status == "active" {
				port, err := c.parsePortFromDNSAddress(payload.DNSAddress)
				if err != nil {
					return
				}
				ch <- cluster.InstanceCredentials{
					UID:      payload.UID,
					Port:     port,
					IPList:   payload.IPList,
					Password: payload.Password,
				}
				break
			}

			res, err = httpClient.Get(fmt.Sprintf("/v1/bdbs/%d", payload.UID), httpclient.HTTPParams{})
			if err != nil {
				c.logger.Error("Failed to make a polling request", err)
			}
		}
	}()
	return ch, nil
}

func (c *apiClient) UpdateDatabase(UID int, params map[string]interface{}) error {
	httpClient := c.httpClient()

	bytes, err := json.Marshal(params)
	if err != nil {
		c.logger.Error("Failed to serialize update parameters", err)
	}

	c.logger.Info("Sending a database update request", lager.Data{
		"UID":        UID,
		"Parameters": params,
	})
	res, err := httpClient.Put(fmt.Sprintf("/v1/bdbs/%d", UID), httpclient.HTTPPayload(bytes))
	if err != nil {
		c.logger.Error("Failed to perform an update request", err, lager.Data{
			"UID": UID,
		})
		return err
	}

	if res.StatusCode != 200 {
		payload, err := c.parseErrorResponse(res)
		if err != nil {
			return err
		}
		err = fmt.Errorf(payload.ErrorMessage)
		c.logger.Error("Failed to update the database", err, lager.Data{
			"UID": UID,
		})
		return err
	}

	c.logger.Info("The database update has been scheduled", lager.Data{
		"UID": UID,
	})
	return nil
}

func (c *apiClient) DeleteDatabase(UID int) error {
	httpClient := c.httpClient()

	res, err := httpClient.Delete(fmt.Sprintf("/v1/bdbs/%d", UID))
	if err != nil {
		c.logger.Error("Failed to perform the database removal request", err, lager.Data{
			"UID": UID,
		})
		return err
	}

	if res.StatusCode != 200 {
		payload, err := c.parseErrorResponse(res)
		if err != nil {
			return err
		}
		err = fmt.Errorf(payload.ErrorMessage)
		c.logger.Error("Failed to delete the database", err)
		return err
	}

	c.logger.Info("The database removal has been scheduled", lager.Data{
		"UID": UID,
	})
	return nil
}

func (c *apiClient) httpClient() httpclient.HTTPClient {
	return httpclient.New(
		c.conf.Redislabs.Auth.Username,
		c.conf.Redislabs.Auth.Password,
		c.conf.Redislabs.Address,
		c.conf.Redislabs.Port,
		c.logger,
	)
}

func (c *apiClient) parseErrorResponse(res *http.Response) (errorResponse, error) {
	payload := errorResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err == nil {
		err = json.Unmarshal(bytes, &payload)
	}
	if err != nil {
		c.logger.Error("Failed to parse the error response payload", err, lager.Data{
			"response": string(bytes),
		})
		err = fmt.Errorf("an unknown server error occurred")
	}
	return payload, err
}

func (c *apiClient) parseStatusResponse(res *http.Response) (statusResponse, error) {
	payload := statusResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err == nil {
		err = json.Unmarshal(bytes, &payload)
	}
	if err != nil {
		c.logger.Error("Failed to parse the status response payload", err)
	}
	return payload, err
}

func (c *apiClient) parsePortFromDNSAddress(address string) (int, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		err := fmt.Errorf("DNS address does not contain port")
		c.logger.Error("Failed to parse the port", err)
		return 0, err
	}
	port, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		c.logger.Error("Failed to parse the port", err)
		return 0, err
	}
	return int(port), nil
}
