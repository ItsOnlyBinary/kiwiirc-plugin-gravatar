package main

import (
	"log"
	"net/http"
)

var router *http.ServeMux
var server *http.Server

func main() {
	router = http.NewServeMux()
	startGravatar(router)

	if config.ListenAddress == "" {
		log.Println("Config error listen_address missing")
		return
	}

	server = &http.Server{
		Addr:    config.ListenAddress,
		Handler: router,
	}
	log.Printf("Listening on %s", config.ListenAddress)
	server.ListenAndServe()
}
