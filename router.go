// Copyright Piero de Salvia.
// All Rights Reserved

package dynaroutes

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/pierods/pluginator"
)

func NewRouterF(interfaceName string, port uint, readTimeout, writeTimeout time.Duration, pluginDir string, consoleHost string, consolePort uint) (*Router, error) {

	router, err := newRouter(interfaceName, port, readTimeout, writeTimeout)
	if err != nil {
		return nil, err
	}
	router.pluginDir = pluginDir
	pluginator, err := pluginator.NewPluginatorF(pluginDir)
	if err != nil {
		return nil, err
	}
	pluginator.SubscribeScan(router.ScanSubscriber)
	pluginator.SubscribeUpdate(router.UpdateSubscriber)
	pluginator.SubscribeAdd(router.AddSubscriber)
	pluginator.SubscribeRemove(router.RemoveSubscriber)

	router.pluginator = pluginator

	log.Println("Router created with plugin dir " + pluginDir)
	go router.startConsole(consoleHost, int(consolePort))
	return router, nil
}

func NewRouterC(interfaceName string, port uint, readTimeout, writeTimeout time.Duration, consulHost string, consulPort uint, consulPrefix string, consoleHost string, consolePort uint) (*Router, error) {

	router, err := newRouter(interfaceName, port, readTimeout, writeTimeout)
	if err != nil {
		return nil, err
	}

	pluginator, err := pluginator.NewPluginatorC(consulHost, int(consulPort), consulPrefix)
	if err != nil {
		return nil, err
	}
	pluginator.SubscribeScan(router.ScanSubscriber)
	pluginator.SubscribeUpdate(router.UpdateSubscriber)
	pluginator.SubscribeAdd(router.AddSubscriber)
	pluginator.SubscribeRemove(router.RemoveSubscriber)

	router.pluginator = pluginator

	log.Println("Router created with consul: "+consulHost, ":", consulPort, ":", consulPrefix)
	go router.startConsole(consoleHost, int(consolePort))

	return router, nil
}

func newRouter(interfaceName string, port uint, readTimeout, writeTimeout time.Duration) (*Router, error) {

	portS := strconv.FormatUint(uint64(port), 10)
	router := Router{}

	server := http.Server{
		Addr: interfaceName + ":" + portS,
		Handler: &MainHandler{
			router: &router,
		},
		ReadTimeout: readTimeout,
		// cannot use WriteTimeout - if a proxy req hangs, ServeHttp will think it's incurring in a write problem and just close the connection,
		// without giving an error message about the proxy request
	}
	router.server = &server
	router.client = http.Client{}
	router.requestTimeout = writeTimeout
	router.prefilterCode = make(map[string]string)
	router.postFilterCode = make(map[string]string)
	return &router, nil
}

func (r *Router) Start() {
	r.pluginator.Start()
	log.Println("Router listening on " + r.server.Addr)
	err := r.server.ListenAndServe()
	log.Println(err)
}

func (r *Router) Shutdown() {
	r.pluginator.Terminate()
	r.server.Shutdown(context.TODO())
}

func (r *Router) AddSubscriber(fileName string, pluginLib *pluginator.PluginContent) {
	r.handlePluginLib(fileName, pluginLib)
}

func (r *Router) ScanSubscriber(pluginNamesAndLibs map[string]*pluginator.PluginContent) {

	for fileName, pluginLib := range pluginNamesAndLibs {
		r.handlePluginLib(fileName, pluginLib)
	}
}

func (r *Router) RemoveSubscriber(fileName string, pluginLib *pluginator.PluginContent) {
	preFilterLib, err := pluginLib.Lib.Lookup("PreFilter")
	if err == nil {
		preFilterPtr, isInstanceOf := preFilterLib.(*PreFilter)
		if isInstanceOf {
			preFilter := *preFilterPtr
			found := false
			where := 0
			for i, rPreFilter := range r.preFilters {
				if preFilter.Name() == rPreFilter.Name() {
					found = true
					where = i
					break
				}
			}
			if found {
				delete(r.prefilterCode, r.preFilters[where].Name())
				r.preFilters[where] = r.preFilters[len(r.preFilters)-1]
				r.preFilters[len(r.preFilters)-1] = nil
				r.preFilters = r.preFilters[:len(r.preFilters)-1]

			}
			sort.Sort(PreFilterByOrder(r.preFilters))
		} else {
			log.Println("file ", fileName, " contains a PreFilter that does not implement dynaroutes.PostFilter")
		}
	}
	postFilterLib, err := pluginLib.Lib.Lookup("PostFilter")
	if err == nil {
		postFilterPtr, isInstanceOf := postFilterLib.(*PostFilter)
		if isInstanceOf {
			postFilter := *postFilterPtr
			found := false
			where := 0
			for i, rPostFilter := range r.postFilters {
				if postFilter.Name() == rPostFilter.Name() {
					found = true
					where = i
					break
				}
			}
			if found {
				delete(r.postFilterCode, r.postFilters[where].Name())
				r.postFilters[where] = r.postFilters[len(r.postFilters)-1]
				r.postFilters[len(r.postFilters)-1] = nil
				r.postFilters = r.postFilters[:len(r.postFilters)-1]
			}
			sort.Sort(PostFilterByOrder(r.postFilters))
		} else {
			log.Println("file ", fileName, " contains a PostFilter that does not implement dynaroutes.PostFilter")
		}
	}
}

func (r *Router) UpdateSubscriber(fileName string, pluginLib *pluginator.PluginContent) {

	r.handlePluginLib(fileName, pluginLib)
}

func (r *Router) handlePluginLib(fileName string, pluginLib *pluginator.PluginContent) {

	preFilterLib, err := pluginLib.Lib.Lookup("PreFilter")
	if err == nil {
		preFilterPtr, isInstanceOf := preFilterLib.(*PreFilter)
		if isInstanceOf {
			preFilter := *preFilterPtr
			found := false
			where := 0
			for i, rPreFilter := range r.preFilters {
				if preFilter.Name() == rPreFilter.Name() {
					found = true
					where = i
					break
				}
			}
			if found {
				r.preFilters[where] = preFilter
				log.Println("Updated pre filter ", preFilter.Name())
			} else {
				r.preFilters = append(r.preFilters, preFilter)
				log.Println("Added pre filter ", preFilter.Name())
			}
			sort.Sort(PreFilterByOrder(r.preFilters))
			r.prefilterCode[preFilter.Name()] = pluginLib.Code
		} else {
			log.Println("file ", fileName, " contains a PreFilter that does not implement dynaroutes.PostFilter")
		}
	}
	postFilterLib, err := pluginLib.Lib.Lookup("PostFilter")
	if err == nil {
		postFilterPtr, isInstanceOf := postFilterLib.(*PostFilter)
		if isInstanceOf {
			postFilter := *postFilterPtr
			found := false
			where := 0
			for i, rPostFilter := range r.postFilters {
				if postFilter.Name() == rPostFilter.Name() {
					found = true
					where = i
					break
				}
			}
			if found {
				r.postFilters[where] = postFilter
				log.Println("Updated post filter ", postFilter.Name())
			} else {
				r.postFilters = append(r.postFilters, postFilter)
				log.Println("Added post filter ", postFilter.Name())
			}
			sort.Sort(PostFilterByOrder(r.postFilters))
			r.postFilterCode[postFilter.Name()] = pluginLib.Code
		} else {
			log.Println("file ", fileName, " contains a PostFilter that does not implement dynaroutes.PostFilter")
		}
	}
}

type PreFilterByOrder []PreFilter

func (a PreFilterByOrder) Len() int           { return len(a) }
func (a PreFilterByOrder) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PreFilterByOrder) Less(i, j int) bool { return a[i].Order() < a[j].Order() }

type PostFilterByOrder []PostFilter

func (a PostFilterByOrder) Len() int           { return len(a) }
func (a PostFilterByOrder) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PostFilterByOrder) Less(i, j int) bool { return a[i].Order() < a[j].Order() }

type MainHandler struct {
	router *Router
}

func (m *MainHandler) handleFilterError(responseWriter http.ResponseWriter, request *http.Request, err error) {
	responseWriter.Header().Set("Content/Type", "text/html")
	responseWriter.WriteHeader(500)
	responseWriter.Write([]byte(err.Error()))
}

func (m *MainHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {

	var route *Route
	var err error
	for _, router := range m.router.preFilters {
		route, err = router.Filter(request)
		if err != nil {
			m.handleFilterError(responseWriter, request, err)
			return
		}
		if route != nil {
			break
		}
	}
	if route != nil {
		portS := strconv.Itoa(route.Port)
		proxyURL, err := url.Parse(route.Scheme + "://" + route.Host + ":" + portS + route.Uri)
		if err != nil {
			responseWriter.Header().Set("Content/Type", "text/html")
			responseWriter.WriteHeader(500)
			responseWriter.Write([]byte(err.Error()))
			return
		}
		proxyReq := new(http.Request)
		*proxyReq = *request
		proxyReq.URL = proxyURL
		proxyReq.RequestURI = ""
		proxyReq.Method = route.Method

		if request.ContentLength > 0 {
			proxyReq.Body = request.Body
		}
		var withDeadline context.Context
		var cancel context.CancelFunc
		if route.Timeout != 0 {
			withDeadline, cancel = context.WithDeadline(request.Context(), time.Now().Add(route.Timeout))
		} else {
			withDeadline, cancel = context.WithDeadline(request.Context(), time.Now().Add(m.router.requestTimeout))
		}
		defer cancel()
		proxyReq = proxyReq.WithContext(withDeadline)
		proxyReq.Header = m.cloneHeader(request.Header)

		proxyResp, err := m.router.client.Do(proxyReq)
		if err != nil {
			responseWriter.Header().Set("Content/Type", "text/html")
			responseWriter.WriteHeader(500)
			responseWriter.Write([]byte(err.Error()))
			return
		}
		defer proxyResp.Body.Close()
		var filteredBody []byte
		for _, postFilter := range m.router.postFilters {
			filteredBody, err = postFilter.Filter(request, proxyReq, proxyResp)
			if err != nil {
				m.handleFilterError(responseWriter, request, err)
				return
			}
			if filteredBody != nil {
				break
			}
		}

		m.copyHeader(responseWriter.Header(), proxyResp.Header)
		if filteredBody != nil {
			responseWriter.Header().Set("Content-Length", strconv.Itoa(len(filteredBody)))
			io.Copy(responseWriter, bytes.NewReader(filteredBody))
			return
		}
		responseWriter.WriteHeader(proxyResp.StatusCode)
		io.Copy(responseWriter, proxyResp.Body)

	} else {
		responseWriter.Header().Set("Content/Type", "text/html")
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte("Not Found"))
		return
	}
}

func (m *MainHandler) copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (m *MainHandler) cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}
