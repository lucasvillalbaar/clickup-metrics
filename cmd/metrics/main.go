package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/api"
)

func main() {

	router := api.Init()
	log.Println("Metrics API")
	err := http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, router))
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
