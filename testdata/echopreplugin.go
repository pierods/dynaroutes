package main

import (
	"context"
	"net/http"

	"github.com/pierods/dynaroutes"
)

type EchoRoutePreFilter struct{}

func (srf *EchoRoutePreFilter) Filter(request *http.Request) (*dynaroutes.Route, error) {

	newCtx := context.WithValue(request.Context(), "filterkey", "this value will be available to other pre filters and post filters")
	*request = *request.WithContext(newCtx)
	return &dynaroutes.Route{
		Scheme: "http",
		Host:   "localhost",
		URI:    request.URL.String(),
		Port:   50000,
		Method: request.Method,
	}, nil
}
func (srf *EchoRoutePreFilter) Order() int {
	return 100
}

func (srf *EchoRoutePreFilter) Description() string {
	return "A simple route filter - forwards to localhost:50000"
}

func (srf *EchoRoutePreFilter) Name() string {
	return "EchoRoutePreFilter"
}

var PreFilter dynaroutes.PreFilter = &EchoRoutePreFilter{}

func main() {}
