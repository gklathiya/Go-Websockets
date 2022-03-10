package main

import (
	"log"
	"net/http"

	"github.com/gklathiya/Go-Websockets/internal/handlers"
)

func main() {

	mux := routes()

	log.Println("Starting Channel Listener")
	go handlers.ListenToWSChannel()

	log.Println("Starting Web server on Port 8080")
	_ = http.ListenAndServe(":8080", mux)
}
