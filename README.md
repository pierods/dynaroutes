# dynaroutes - a reverse proxy for dynamic routing, with scripted Go

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![](https://godoc.org/github.com/pierods/dynaroutes?status.svg)](http://godoc.org/github.com/pierods/dynaroutes)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierods/dynaroutes)](https://goreportcard.com/report/github.com/pierods/dynaroutes)
[![Build Status](https://travis-ci.org/pierods/dynaroutes.svg?branch=master)](https://travis-ci.org/pierods/dynaroutes)

dynaroutes is a reverse proxy, inspired by [zuul](https://github.com/Netflix/zuul). It features:

+ dynamic filters - written in Go and compiled on the fly. 
+ directory-based or consul-based filter loading/unloading/editing

## Use cases
dynaroutes is useful for:
+ A/B testing at full speed - only compiled code is used
+ Quick scaffolding of microservice architectures
+ On-the-fly routing changes, hot reload of filters


## What is a filter

A filter is a small program that is applied either at the moment when a request comes into the reverse proxy ("pre filters") or after 
a proxy request has returned ("post filters") - but before the final response is sent out.
Many pre and post filters can be added to dynaroutes. Every filter has got an order, and they are apllied in that order.

### pre filters

Pre filters have this interface:

```Go
    type PreFilter interface {
        Name() string
        Order() int
        Description() string
        //Filter returns a Route in case it decides to route the request, nil otherwise.
        Filter(request *http.Request) (*Route, error)
    }
```
Pre filters are able to determine where a request will be proxied to. They have the complete, original request to make that decision:

```Go
// Route is returned by pre filters to set the reverse proxy request parameters
    type Route struct {
        Scheme  string
        Host    string
        URI     string
        Port    int
        Method  string
        Timeout time.Duration
}
```
Pre filters are applied (Filter() is called) until a filter returns a route. Then the router stops applying filters and forwards the request 
to the desired endpoint, as described by the returned Route. A Route can set the Scheme/host/URI/Port/Method/Timeout of a request.

### post filters

After the proxy request has returned, post filters are called. A post filter has this interface:

```Go
    //PostFilter is applied before a response is sent back to the client. Many post filter can be defined, and they are applied in the order defined by Order(), however as soon as a non-nil []byte is returned, the router stops applying post filters.
    type PostFilter interface {
        Name() string
        Order() int
        Description() string
        //Filter can return nil, or a []byte, in which case that's what will be sent back to the client
        Filter(request, proxyRequest *http.Request, proxyResponse *http.Response) ([]byte, error)
    }

```
