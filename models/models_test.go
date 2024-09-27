package models

import (
	"net/url"
	"testing"
)

func TestLoadBalancer(t *testing.T) {
	lb := &LoadBalancer{}
	servers := []*Server{
		{URL: &url.URL{Host: "server1"}, IsHealthy: true},
		{URL: &url.URL{Host: "server2"}, IsHealthy: false},
		{URL: &url.URL{Host: "server3"}, IsHealthy: true},
	}

	server := lb.getNextServer(servers)
	if server.URL.Host != "server1" {
		t.Errorf("Expected server1, got %s", server.URL.Host)
	}

	server = lb.getNextServer(servers)
	if server.URL.Host != "server3" {
		t.Errorf("Expected server3, got %s", server.URL.Host)
	}
}
