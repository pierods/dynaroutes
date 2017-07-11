// Copyright Piero de Salvia.
// All Rights Reserved

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pierods/dynaroutes"
)

var (
	port          = flag.Uint("port", 8080, "port is the port the router will listen on. Defaults to 8080")
	interfaceName = flag.String("interface", "localhost", "interface is the name of the interface the router will listen on. Defaults to localhost")
	readTimeout   = flag.Uint("readtimeout", 10, "readtimeout is the read timeout in seconds. Defaults to 10.")
	writeTimeout  = flag.Uint("writetimeout", 10, "writetimeout is the write timeout in seconds. Defaults to 10")
	filterDir     = flag.String("filterdir", "", "filterdir is the directory containing the source code of the filters")
	consulMode    = flag.Bool("consul", false, "enables consul mode")
	consulPort    = flag.Uint("consulport", 8500, "consulport is the port consul is listening on")
	consulHost    = flag.String("consulhost", "localhost", "consulhost is the host consul is listening on")
	consulPrefix  = flag.String("consulprefix", "com.github.pierods.dynaroutes.filters", "consulprefix is the prefix of the plugin keys")
	consoleHost   = flag.String("consolehost", "localhost", "host for console server")
	consolePort   = flag.Uint("consoleport", 30000, "port to use for the http console")
)

func main() {

	flag.Parse()

	errF := func(msg string) {
		fmt.Println(msg)
		os.Exit(-1)
	}

	if *port == uint(0) {
		errF("port cannot be 0")
	}
	if *interfaceName == "" {
		errF("interfacename cannot be empty")
	}
	if *readTimeout == 0 {
		errF("readtimeout cannot be 0")
	}
	if *writeTimeout == 0 {
		errF("writetimeout cannot be 0")
	}

	if *filterDir == "" {
		errF("filterdir cannot be blank")
	}
	var err error
	var router *dynaroutes.Router
	if *consulMode {
		router, err = dynaroutes.NewRouterC(*interfaceName, *port, time.Duration(*readTimeout)*time.Second, time.Duration(*writeTimeout)*time.Second, *consulHost, *consulPort, *consulPrefix, *consoleHost, *consolePort)
	} else {
		router, err = dynaroutes.NewRouterF(*interfaceName, *port, time.Duration(*readTimeout)*time.Second, time.Duration(*writeTimeout)*time.Second, *filterDir, *consoleHost, *consolePort)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = router.Start()
	if err != nil {
		errF(err.Error())
	}
}
