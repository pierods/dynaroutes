// Copyright Piero de Salvia.
// All Rights Reserved

package dynaroutes

import (
	"net/http"
	"time"

	"github.com/pierods/pluginator"
)

type Router struct {
	preFilters     []PreFilter
	prefilterCode  map[string]string
	postFilters    []PostFilter
	postFilterCode map[string]string
	server         *http.Server
	client         http.Client
	pluginDir      string
	pluginator     *pluginator.Pluginator
	requestTimeout time.Duration
}

type Route struct {
	Scheme  string
	Host    string
	Uri     string
	Port    int
	Method  string
	Timeout time.Duration
}

type FilterBase interface {
	Name() string
	Order() int
	Description() string
}

type PreFilter interface {
	FilterBase
	/*
		Filter returns a Route in case it decides to route the request, nil otherwise.
	*/
	Filter(request *http.Request) (*Route, error)
}

type PostFilter interface {
	FilterBase
	/*
		Filter can return nil, or a *bytes.Buffer, in which case that's what will be sent back to the client
	*/
	Filter(request, proxyRequest *http.Request, proxyResponse *http.Response) ([]byte, error)
}
