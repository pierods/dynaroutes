// Copyright Piero de Salvia.
// All Rights Reserved

package dynaroutes

import (
	"io"
	"net/http"
	"time"

	"github.com/pierods/pluginator"
)

// Router is lib's entry point
type Router struct {
	preFilters        []PreFilter
	prefilterCode     map[string]string
	postFilters       []PostFilter
	postFilterCode    map[string]string
	routingFilters    []RoutingFilter
	routingFilterCode map[string]string
	server            *http.Server
	client            http.Client
	pluginDir         string
	pluginator        *pluginator.Pluginator
	requestTimeout    time.Duration
}

// Route is returned by pre filters to set the reverse proxy request parameters
type Route struct {
	Scheme  string
	Host    string
	URI     string
	Port    int
	Method  string
	Timeout time.Duration
	Body    io.ReadCloser
}

// ResultMsg is the result of an http call. or an error
type ResultMsg struct {
	Response *http.Response
	Err      error
}

// RouteTasks has two channels, one for sending a Route to be proxies, one for receiving the result of the routing
type RouteTasks struct {
	// this notifies the router that filter is ready to have a request sent out. Channel is necessary because if a routing filter decides
	//to compose, it might need the result(s) of a previous proxying to issue the next proxying request
	Send    <-chan *Route
	Receive chan<- *ResultMsg
}

// FilterBase has the base methods for filters
type FilterBase interface {
	Name() string
	Order() int
	Description() string
}

// A PreFilter is applied to a request before proxying. Many pre filters cam be defined, and they are applied in the order defined by Order(), however as soon as a non-nil route is returned, the router stops applying pre filters.
type PreFilter interface {
	FilterBase
	//Filter has full access to the Request and can modify it.
	Filter(request *http.Request) error
}

/*
RoutingFilter is the filter type for routing
*/
type RoutingFilter interface {
	FilterBase
	/*
		Routes returns a *RouteTask, or nil if it decides not to handle the request. When the code is ready to send
		the final result to the router, it should push it into endResult. Notice that  request also carries the context.
	*/
	Filter(request *http.Request, endResult chan<- *ResultMsg) (*RouteTasks, error)
}

//PostFilter is applied before a response is sent back to the client. Many post filter can be defined, and they are applied in the order defined by Order(), however as soon as a non-nil []byte is returned, the router stops applying post filters.
type PostFilter interface {
	FilterBase
	/*
		Filter can return nil, or a []byte, in which case that's what will be sent back to the client and no more filters will be applied. Notice that
		request also carries the context.
	*/
	Filter(request *http.Request, response *http.Response) ([]byte, error)
}
