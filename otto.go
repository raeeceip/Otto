package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func checkPrerequisites() tea.Msg {
	// Check if config.json exists
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		return prerequisitesCheckedMsg{err: fmt.Errorf("config.json not found")}
	}

	// Load and parse config
	configFile, err := os.Open("config.json")
	if err != nil {
		return prerequisitesCheckedMsg{err: fmt.Errorf("error opening config file: %v", err)}
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return prerequisitesCheckedMsg{err: fmt.Errorf("error decoding config file: %v", err)}
	}

	// Check if load balancer port is available
	if !isPortAvailable(config.Port) {
		return prerequisitesCheckedMsg{err: fmt.Errorf("load balancer port %s is not available", config.Port)}
	}

	// Check if all server ports are available
	unavailablePorts := []string{}
	for _, serverURL := range config.Servers {
		u, err := url.Parse(serverURL)
		if err != nil {
			return prerequisitesCheckedMsg{err: fmt.Errorf("invalid server URL: %s", serverURL)}
		}
		if !isPortAvailable(u.Port()) {
			unavailablePorts = append(unavailablePorts, u.Port())
		}
	}

	if len(unavailablePorts) > 0 {
		return prerequisitesCheckedMsg{err: fmt.Errorf("the following ports are not available: %v", unavailablePorts)}
	}

	return prerequisitesCheckedMsg{err: nil}
}

func isPortAvailable(port string) bool {
	port = strings.TrimPrefix(port, ":")
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func loadData() tea.Msg {
	// Load config and initialize servers
	configFile, _ := os.Open("config.json")
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	decoder.Decode(&config)

	healthCheckInterval, _ := time.ParseDuration(config.HealthCheckInterval)

	Servers = make([]*Server, len(config.Servers))
	for i, serverURL := range config.Servers {
		url, _ := url.Parse(serverURL)
		Servers[i] = &Server{URL: url, IsHealthy: false}
	}

	// Perform initial health check
	var wg sync.WaitGroup
	for _, server := range Servers {
		wg.Add(1)
		go func(s *Server) {
			defer wg.Done()
			healthCheck(s, healthCheckInterval)
		}(server)
	}
	wg.Wait()

	return dataLoadedMsg{servers: Servers}
}

func healthCheck(s *Server, healthCheckInterval time.Duration) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Head(s.URL.String())
	s.Mutex.Lock()
	if err != nil || res.StatusCode != http.StatusOK {
		s.IsHealthy = false
	} else {
		s.IsHealthy = true
	}
	s.Mutex.Unlock()
}

func main() {
	// Create a channel to signal when the TUI is done
	tuiDone := make(chan struct{})

	go func() {
		runTUI()
		close(tuiDone)
	}()

	// Wait for TUI to finish (this happens when prerequisites are checked)
	<-tuiDone

	// Load config and start load balancer after TUI confirms prerequisites
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}

	lb := &LoadBalancer{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.getNextServer(Servers)
		if server == nil {
			http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
			return
		}
		w.Header().Add("X-Forwarded-Server", server.URL.String())
		server.ReverseProxy().ServeHTTP(w, r)
	})

	port := strings.TrimPrefix(config.Port, ":")
	log.Printf("Starting Otto load balancer on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
