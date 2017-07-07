package dynaroutes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (r *Router) startConsole(host string, port int) {

	portS := strconv.Itoa(port)

	server := http.Server{
		Addr: host + ":" + portS,
		Handler: &ConsoleHandler{
			router: r,
		},
	}

	log.Println("Web console started on ", host, ":", port)
	err := server.ListenAndServe()
	log.Println(err)
}

func readFile(name string) ([]byte, error) {

	goPath := os.Getenv("GOPATH")
	filePath := goPath + "/src/github.com/pierods/dynaroutes/assets" + name

	f, fErr := os.Open(filePath)
	defer f.Close()

	if fErr != nil {
		return nil, fErr
	}
	bytes, fErr := ioutil.ReadAll(f)

	if fErr != nil {
		return nil, fErr
	}
	return bytes, nil
}

type ConsoleHandler struct {
	router *Router
}

func (ch *ConsoleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/pre") {
		pres := ch.router.preList()
		json, err := json.Marshal(pres)
		if err != nil {
			rw.WriteHeader(500)
			fmt.Fprint(rw, err)
			return
		}
		rw.Write(json)
		return
	}

	if strings.HasPrefix(path, "/post") {
		posts := ch.router.postList()
		json, err := json.Marshal(posts)
		if err != nil {
			rw.WriteHeader(500)
			fmt.Fprint(rw, err)
			return
		}
		rw.Write(json)
		return
	}

	if path == "/" {
		path = "/index.html"
	}
	page, err := readFile(path)
	rw.Header().Set("Content/Type", "text/html")
	if err != nil {
		rw.WriteHeader(404)
		fmt.Fprint(rw, err)
	}
	rw.Write(page)
}

type FilterItem struct {
	Name        string `json:"name"`
	Order       int    `json:"order"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

func (r *Router) preList() []FilterItem {

	var items []FilterItem
	for _, pre := range r.preFilters {
		item := FilterItem{
			Name:        pre.Name(),
			Order:       pre.Order(),
			Description: pre.Description(),
			Code:        r.prefilterCode[pre.Name()],
		}
		items = append(items, item)
	}

	return items
}

func (r *Router) postList() []FilterItem {

	var items []FilterItem
	for _, post := range r.postFilters {
		item := FilterItem{
			Name:        post.Name(),
			Order:       post.Order(),
			Description: post.Description(),
			Code:        r.postFilterCode[post.Name()],
		}
		items = append(items, item)
	}

	return items
}
