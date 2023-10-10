package main

import (
	"log"
	"net/http"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/api"
)

func main() {

	router := api.Init()
	log.Println("Metrics API")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
