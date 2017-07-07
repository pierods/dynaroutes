package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pierods/dynaroutes"
)

type SimplePostFilter struct{}

func (spf *SimplePostFilter) Filter(request, proxyRequest *http.Request, proxyResponse *http.Response) ([]byte, error) {

	log.Print("post filter: original request:", request.URL.String())
	log.Println(" was forwarded to:", proxyRequest.URL.String())

	respB, err := ioutil.ReadAll(proxyResponse.Body)
	if err != nil {
		return nil, err
	}

	return append(respB, []byte(" - filtered")...), nil

	return nil, nil

}
func (spf *SimplePostFilter) Order() int {
	return 10
}

func (spf *SimplePostFilter) Description() string {
	return "A simple post filter - logs basic info"
}

func (spf *SimplePostFilter) Name() string {
	return "SimplePostFilter"
}

var PostFilter dynaroutes.PostFilter = &SimplePostFilter{}

func main() {}
