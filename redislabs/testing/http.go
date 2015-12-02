package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type (
	HTTPProxy interface {
		URL() string
		RegisterEndpoints(endpoints []Endpoint)
		Close()
	}
	Endpoint struct {
		URL      string
		Response interface{}
	}

	httpProxy struct {
		Mux    *http.ServeMux
		Server *httptest.Server
	}
)

// NewHTTPProxy creates a proxy that mocks HTTP requests to the registered endpoints.
// Endpoints can be registered via the RegisterEndpoints method.
// The proxy should be shutdown via Close.
func NewHTTPProxy() *httpProxy {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	return &httpProxy{
		Mux:    mux,
		Server: server,
	}
}

func (p *httpProxy) URL() string {
	return p.Server.URL
}

func (p *httpProxy) RegisterEndpoints(endpoints []Endpoint) {
	for _, e := range endpoints {
		e := e // escape the closure
		p.Mux.HandleFunc(e.URL, func(w http.ResponseWriter, r *http.Request) {
			js, err := json.Marshal(e.Response)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		})
	}
}

func (p *httpProxy) Close() {
	p.Server.Close()
}
