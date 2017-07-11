// Copyright Piero de Salvia.
// All Rights Reserved

/*
moved test to this package because of https://github.com/golang/go/issues/17928
*/
package dynaroutes_integration

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pierods/dynaroutes"
)

var testDataDir string

func init() {

	goPath := os.Getenv("GOPATH")
	testDataDir = goPath + "/src/github.com/pierods/dynaroutes/testdata"
}

func TestHeaders(t *testing.T) {

	client := http.Client{}

	req, err := http.NewRequest("GET", "http://localhost:30000", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Must", "gothrough")
	req.Header.Add("This", "too")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("Must") != "gothrough" || resp.Header.Get("This") != "too" {
		t.Fatal("Should be able to carry headers through")
	}
}

func TestBody(t *testing.T) {

	client := http.Client{}

	body := "Body"
	req, err := http.NewRequest("GET", "http://localhost:30000", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	buffer := new(bytes.Buffer)
	io.Copy(buffer, resp.Body)
	responseBody := buffer.String()
	t.Log(body)

	if responseBody != body+" - filtered" {
		t.Fatal("Should be able to carry body through and filter it")
	}
}

func TestTimeouts(t *testing.T) {

	client := http.Client{}

	req, err := http.NewRequest("GET", "http://localhost:30000/timeout", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if resp.StatusCode == 200 {
		t.Fatal("Should err out on a timeout")
	}

	client = http.Client{}

	req, err = http.NewRequest("GET", "http://localhost:30000/timeoutplugin", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = client.Do(req)
	if resp.StatusCode == 200 {
		t.Fatal("Should err out on a timeout")
	}

}

func TestMain(m *testing.M) {

	listener, err := net.Listen("tcp", ":50000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	server := http.Server{
		Handler: &EchoHandler{},
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	defer server.Close()

	router, err := dynaroutes.NewRouterF("localhost", 30000, 5*time.Second, 5*time.Second, testDataDir, "localhost", 31000)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	go router.Start()

	defer router.Shutdown()
	time.Sleep(60 * time.Second)
	retCode := m.Run()
	os.Exit(retCode)

}

type EchoHandler struct{}

func (e *EchoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.URL.String(), "timeout") {
		time.Sleep(1000 * time.Minute)
		return
	}
	copyHeader(rw.Header(), r.Header)
	rw.Header().Set("Request URL", r.URL.String())
	rw.Header().Set("Request method", r.Method)
	rw.Header().Set("Request Content-Length", strconv.FormatInt(r.ContentLength, 10))
	rw.WriteHeader(200)
	if r.ContentLength > 0 {
		defer r.Body.Close()
		io.Copy(rw, r.Body)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
