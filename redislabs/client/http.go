package client

import (
	"github.com/pivotal-golang/lager"
	"net/http"
	"net/url"
)

var defaultClient = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

type httpParams map[string]string
type httpPayload []byte

type httpClient struct {
	password string
	username string
	address string
	port    int
	logger  *lager.Logger
	client  *http.Client
}

func (c *httpClient) performRequest(verb string, path string, params httpParams, payload httpPayload) (http.Response, error) {
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
		return nil, err
	}
	req.SetBasicAuth(c.username, c.password)
	return c.httpClient.Do(req)
}

func (c *RedislabsClient) put(path string, payload httpPayload) (*http.Response, error) {
	response, err := c.performRequest("PUT", path, httpParams{}, payload)
	if err != nil {
		Logger.Fatal("Performing PUT request", err, lager.Data{
			"endoint": endpoint, "payload": payload
		})
		return nil, err
	}
	return response, nil
}

func (c *RedislabsClient) post(endpoint string, payload httpPayload) (*http.Response, error) {
	response, err := c.performRequest("POST", path, httpParams{}, payload)
	if err != nil {
		Logger.Fatal("Performing POST request", err, lager.Data{
			"endoint": endpoint, "payload": payload
		})
		return nil, err
	}
	return response, nil
}

func (c *httpClient) get(path string, params httpParams) (*http.Response, error) {
	response, err := c.performRequest("GET", path, params, httpPayload{})
	if err != nil {
		Logger.Fatal("Performing GET request", err, lager.Data{
			"endoint": endpoint, "payload": payload
		})
		return nil, err
	}
	return response, nil
}

// TODO: remove following comment
// playground https://play.golang.org/p/juw99Hp9yF
func (c *httpClient) buildFullRequestURL(path string, params httpParams) {
	base_url, _ := url.Parse(c.address)
	endpoint, _ := base_url.Parse(path)
	query := endpoint.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

res, err := httpClient.get()
client.parseResponse(res, Database)

// parse the response
func parseJSONResponse(response *http.Response, result interface{}) error {
	//read the response
	bytes, err := ioutil.ReadAll(response.Body)
	if error != nil {
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
		return errors.New(fmt.Sprintf("Bad Request: %s", string(bytes)))
	}

	return nil
}

func newHTTPClient(username string, password string, address string, port int, logger *lager.Logger) httpClient {
	logger.Info("Creating new http client", lager.Data{address: address, port: port})
	return &httpClient{
		address: address,
		port: port,
		logger: logger,
		client: defaultClient
	}
}
