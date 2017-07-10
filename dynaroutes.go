// Copyright Piero de Salvia.
// All Rights Reserved

package dynaroutes

import (
	"net/http"
	"time"

	"github.com/pierods/pluginator"
)

// Router is lib's entry point
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

// Route is returned by pre filters to set the reverse proxy request parameters
type Route struct {
	Scheme  string
	Host    string
	URI     string
	Port    int
	Method  string
	Timeout time.Duration
}

type filterBase interface {
	Name() string
	Order() int
	Description() string
}

// A PreFilter is applied to a request before proxying. Many pre filters cam be defined, and they are applied in the order defined by Order(), however as soon as a non-nil route is returned, the router stops applying pre filters.
type PreFilter interface {
	filterBase
	//Filter returns a Route in case it decides to route the request, nil otherwise.
	Filter(request *http.Request) (*Route, error)
}

//PostFilter is applied before a response is sent back to the client. Many post filter can be defined, and they are applied in the order defined by Order(), however as soon as a non-nil []byte is returned, the router stops applying post filters.
type PostFilter interface {
	filterBase
	//Filter can return nil, or a []byte, in which case that's what will be sent back to the client
	Filter(request, proxyRequest *http.Request, proxyResponse *http.Response) ([]byte, error)
}
