package httpclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		c.logger.Fatal("Performing PUT request", err, lager.Data{
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
		c.logger.Fatal("Performing POST request", err, lager.Data{
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
		c.logger.Fatal("Performing GET request", err, lager.Data{
			"endoint": endpoint,
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
		"performing-http-request",
		lager.Data{
			"verb":    verb,
			"path":    path,
			"params":  params,
			"payload": payload,
		},
	)
	// TODO: validate inputs (for instance verb)
	requestURL := c.buildFullRequestURL(path, params)
	req, err := http.NewRequest(verb, requestURL, nil)
	if err != nil {
		return &http.Response{}, err
	}
	req.SetBasicAuth(c.username, c.password)
	return c.client.Do(req)
}

// parse the response
func parseJSONResponse(response *http.Response, result interface{}) error {
	//read the response
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// close the body when done reading
	defer response.Body.Close()

	//parse JSON
	err = json.Unmarshal(bytes, result)
	if err != nil {
		return err
	}

	//check whether the response is a bad request
	if response.StatusCode == 400 {
		return fmt.Errorf("Bad Request: %s", string(bytes))
	}

	return nil
}
