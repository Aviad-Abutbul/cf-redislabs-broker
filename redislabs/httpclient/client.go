package httpclient

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pivotal-golang/lager"
)

type (
	HTTPParams  map[string]string
	HTTPPayload []byte
	HTTPClient  interface {
		Get(endpoint string, params HTTPParams) (*http.Response, error)
		Post(endpoint string, payload HTTPPayload) (*http.Response, error)
		Put(endpoint string, payload HTTPPayload) (*http.Response, error)
		Delete(endpoint string) (*http.Response, error)
	}

	httpClient struct {
		password string
		username string
		address  string
		port     int
		logger   lager.Logger
		client   *http.Client
	}
)

var defaultClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // TODO: make it configurable
		Proxy:           http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

// New returns a client that implements HTTPClient interface.
func New(username string, password string, address string, port int, logger lager.Logger) *httpClient {
	logger.Info("Creating new http client", lager.Data{"address": address, "port": port})
	return &httpClient{
		username: username,
		password: password,
		address:  address,
		port:     port,
		logger:   logger,
		client:   defaultClient,
	}
}

func (c *httpClient) Put(endpoint string, payload HTTPPayload) (*http.Response, error) {
	response, err := c.performRequest("PUT", endpoint, HTTPParams{}, payload)
	if err != nil {
		c.logger.Error("Performing PUT request", err, lager.Data{
			"endoint": endpoint,
			"payload": payload,
		})
		return nil, err
	}
	return response, nil
}

func (c *httpClient) Post(endpoint string, payload HTTPPayload) (*http.Response, error) {
	response, err := c.performRequest("POST", endpoint, HTTPParams{}, payload)
	if err != nil {
		c.logger.Error("Performing POST request", err, lager.Data{
			"endoint": endpoint,
			"payload": payload,
		})
		return nil, err
	}
	return response, nil
}

func (c *httpClient) Get(endpoint string, params HTTPParams) (*http.Response, error) {
	response, err := c.performRequest("GET", endpoint, params, HTTPPayload{})
	if err != nil {
		c.logger.Error("Performing GET request", err, lager.Data{
			"endoint": endpoint,
		})
		return nil, err
	}
	return response, nil
}

func (c *httpClient) Delete(endpoint string) (*http.Response, error) {
	response, err := c.performRequest("DELETE", endpoint, HTTPParams{}, HTTPPayload{})
	if err != nil {
		c.logger.Error("Failed to perform DELETE request", err, lager.Data{
			"endpoint": endpoint,
		})
		return nil, err
	}
	return response, nil
}

func (c *httpClient) buildFullRequestURL(path string, params HTTPParams) string {
	baseURL, _ := url.Parse(c.address)
	endpoint, _ := baseURL.Parse(path)
	query := endpoint.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

func (c *httpClient) performRequest(verb string, path string, params HTTPParams, payload HTTPPayload) (*http.Response, error) {
	c.logger.Info(
		"Preparing to perform a request",
		lager.Data{
			"verb":    verb,
			"path":    path,
			"params":  params,
			"payload": payload,
		},
	)
	requestURL := c.buildFullRequestURL(path, params)
	req, err := http.NewRequest(verb, requestURL, bytes.NewReader(payload))
	if err != nil {
		return &http.Response{}, err
	}
	req.SetBasicAuth(c.username, c.password)
	return c.client.Do(req)
}
