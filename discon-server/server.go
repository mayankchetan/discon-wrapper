package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

const program = "discon-server"
const version = "v0.1.0"

// GH-Cp gen: Using an int for debug levels instead of boolean
var debugLevel int = 0
var port = 8080

func main() {

	// Display program and version
	log.Printf("Started %s %s", program, version)

	flag.IntVar(&port, "port", 8080, "Port to listen on")
	// GH-Cp gen: Updated to use debug levels
	flag.IntVar(&debugLevel, "debug", 0, "Debug level: 0=disabled, 1=basic info, 2=verbose with payloads")
	flag.Parse()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, debugLevel)
	})

	// Start server
	log.Printf("Listening on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
