package main

import (
	"log"
	"net/http"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/api"
)

func main() {

	router := api.Init()
	// Iniciar el servidor en el puerto 8080
	log.Println("Server listening on port 8080...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
