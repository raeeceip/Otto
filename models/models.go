package models

import (
	"net/http/httputil"
	"net/url"
	"sync"
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

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
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

var Servers []*Server // Global variable to store servers
