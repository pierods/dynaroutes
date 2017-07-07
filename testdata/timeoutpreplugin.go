package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/pierods/dynaroutes"
)

type EchoRoutePreFilter struct{}

func (srf *EchoRoutePreFilter) Filter(request *http.Request) (*dynaroutes.Route, error) {

	if strings.Contains(request.URL.String(), "timeoutplugin") {
		return &dynaroutes.Route{
			Scheme:  "http",
			Host:    "localhost",
			URI:     request.URL.String(),
			Port:    50000,
			Method:  request.Method,
			Timeout: 1 * time.Second,
		}, nil
	}
	return nil, nil

}
func (srf *EchoRoutePreFilter) Order() int {
	return 10
}

func (srf *EchoRoutePreFilter) Description() string {
	return ""
}

func (srf *EchoRoutePreFilter) Name() string {
	return "TimeoutPrefilter"
}

var PreFilter dynaroutes.PreFilter = &EchoRoutePreFilter{}

func main() {}
