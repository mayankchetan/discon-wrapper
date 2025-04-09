package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

const program = "discon-server"
const version = "v0.1.0"

var debug = false
var port = 8080

func main() {

	// Display program and version
	log.Printf("Started %s %s", program, version)

	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.Parse()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, debug)
	})

	// Start server
	log.Printf("Listening on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
