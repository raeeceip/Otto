package main

import (
	"github.com/charmbracelet/bubbletea"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"
)

type Config struct {
	Port                string   `json:"port"`
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []string `json:"servers"`
}

type Server struct {
	URL       *url.URL
	IsHealthy bool
	Mutex     sync.Mutex
}

type LoadBalancer struct {
	Current int
	Mutex   sync.Mutex
}

func (lb *LoadBalancer) getNextServer(servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	for i := 0; i < len(servers); i++ {
		idx := lb.Current % len(servers)
		nextServer := servers[idx]
		lb.Current++

		nextServer.Mutex.Lock()
		isHealthy := nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}
	}

	return nil
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func healthCheck(s *Server, healthCheckInterval time.Duration) {
	for range time.Tick(healthCheckInterval) {
		res, err := http.Head(s.URL.String())
		s.Mutex.Lock()
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Printf("%s is down\n", s.URL)
			s.IsHealthy = false
		} else {
			s.IsHealthy = true
		}
		s.Mutex.Unlock()
	}
}

func main() {
	go runTUI()
	configFile, err := os.Open("config.json")
		log.Fatal("Error opening config file:", err)
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
		log.Fatal("Error decoding config file:", err)
	}

	healthCheckInterval, err := time.ParseDuration(config.HealthCheckInterval)
		log.Fatal("Invalid health check interval:", err)
	}

	servers := make([]*Server, len(config.Servers))
	for i, serverURL := range config.Servers {
		url, err := url.Parse(serverURL)
			log.Fatal("Invalid server URL:", err)
		}
		servers[i] = &Server{URL: url, IsHealthy: true}
		go healthCheck(servers[i], healthCheckInterval)
	}

	lb := &LoadBalancer{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.getNextServer(servers)
		if server == nil {
			http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
			return
		}
		w.Header().Add("X-Forwarded-Server", server.URL.String())
		server.ReverseProxy().ServeHTTP(w, r)
	})

	go func() {
		if err := http.ListenAndServe(config.Port, nil); err != nil {
			log.Fatalf("Error starting load balancer: %s\n", err.Error())
		}
	}()
	}
}
