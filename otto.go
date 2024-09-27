package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"otto/models"
	"otto/ui"
)

func main() {
	go ui.RunTUI()

	// Load config and start load balancer after TUI confirms prerequisites
	<-time.After(time.Second) // Give TUI time to start

	configFile, _ := os.Open("config.json")
	defer configFile.Close()

	var config models.Config
	decoder := json.NewDecoder(configFile)
	decoder.Decode(&config)

	lb := &models.LoadBalancer{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.GetNextServer(models.Servers)
		if server == nil {
			http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
			return
		}
		w.Header().Add("X-Forwarded-Server", server.URL.String())
		server.ReverseProxy().ServeHTTP(w, r)
	})

	log.Println("Starting Otto load balancer on port", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
