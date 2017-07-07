advantages:

- can read code ON SERVER (show code in web ctrl panel)
- same speed as compiled
- can travel easily over protocols
- must not learn new language
- go used by devops


Filter Interface

Filters must extend ZuulFilter and implement the following methods:

String filterType();

int filterOrder();

boolean shouldFilter();

Object run();


To Run, or Maybe Not

The method shouldFilter() returns a boolean indicating if the Filter should run or not.
Ordering

The method filterOrder() returns an int describing the order that the Filter should run in relative to others.

Filter Types


Zuul's primary request lifecycle consists of "pre", "routing", and "post" phases, in that order. All filters with these types are run for every request.


========================================================================================================================================
"pre" examples

 boolean shouldFilter() {

        if ("true".equals(RequestContext.currentContext.getRequest().getParameter(debugParameter.get()))) return true;
        return routingDebug.get();

    }

-----------------------------------------
boolean shouldFilter() {
        return RequestContext.getCurrentContext().get("ErrorHandled") == null
}

-----------------------------------------

debug Filter
	RequestContext ctx = RequestContext.getCurrentContext()
	ctx.setDebugRouting(true)
	ctx.setDebugRequest(true)

-----------------------------------------
debug Request:
	HttpServletRequest req = RequestContext.currentContext.request as HttpServletRequest // request is available
	// it means log("REQUEST::") etc
	Debug.addRequestDebug("REQUEST:: " + req.getScheme() + " " + req.getRemoteAddr() + ":" + req.getRemotePort())
-----------------------------------------

predecoration (whatever it means)

RequestContext ctx = RequestContext.getCurrentContext()

// sets origin (origin is the url the reverse proxy will use)
ctx.setRouteHost(new URL("http://httpbin.org"));
// sets custom header to send to the origin
ctx.addOriginResponseHeader("cache-control", "max-age=3600");

========================================================================================================================================
route examples: they implement an actual reverse proxy. the origin can be set, for example, in a pre filter, under ""setRouteHost"

========================================================================================================================================
post examples:

boolean shouldFilter() {
        return !RequestContext.getCurrentContext().getZuulResponseHeaders().isEmpty() ||
                RequestContext.getCurrentContext().getResponseDataStream() != null ||
                RequestContext.getCurrentContext().responseBody != null
}

-----------------------------------------

Object run() {
        addResponseHeaders()
        writeResponse()
}


Object run() {
        dumpRoutingDebug()
        dumpRequestDebug()
    }

    public void dumpRequestDebug() {
        List<String> rd = (List<String>) RequestContext.getCurrentContext().get("requestDebug");
        rd?.each {
            println("REQUEST_DEBUG::${it}");
        }
    }

    public void dumpRoutingDebug() {
        List<String> rd = (List<String>) RequestContext.getCurrentContext().get("routingDebug");
        rd?.each {
            println("ZUUL_DEBUG::${it}");
        }
    }
