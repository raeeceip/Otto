package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "sync"
    "time"

    "otto/models"
    "otto/ui"
)

type serverInfo struct {
    server  *models.Server
    lastUsed time.Time
    connections int
}

func main() {
    go ui.RunTUI()

    // Load config and start load balancer after TUI confirms prerequisites
    <-time.After(time.Second) // Give TUI time to start

    configFile, _ := os.Open("config.json")
    defer configFile.Close()

    var config models.Config
    decoder := json.NewDecoder(configFile)
    decoder.Decode(&config)

    lb := &models.LoadBalancer{
        Servers:   make([]*models.Server, 0),
        serverInfo: make(map[*models.Server]*serverInfo),
    }

    var wg sync.WaitGroup

    // Start health check goroutines
    for _, server := range config.Servers {
        wg.Add(1)
        go func(s *models.Server) {
            defer wg.Done()
            for {
                if !s.IsHealthy() {
                    lb.removeServer(s)
                    log.Printf("Server %s is unhealthy\n", s.URL)
                }
                time.Sleep(time.Second * 10) // Adjust health check interval as needed
            }
        }(server)
    }

    // Add initial servers
    for _, server := range config.Servers {
        lb.addServer(server)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        server := lb.getNextServer()
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

    wg.Wait()
}

func (lb *models.LoadBalancer) addServer(server *models.Server) {
    lb.lock.Lock()
    defer lb.lock.Unlock()

    lb.Servers = append(lb.Servers, server)
    lb.serverInfo[server] = &serverInfo{
        server:      server,
        lastUsed:     time.Now(),
        connections: 0,
    }
}

func (lb *models.LoadBalancer) removeServer(server *models.Server) {
    lb.lock.Lock()
    defer lb.lock.Unlock()

    for i, s := range lb.Servers {
        if s == server {
            lb.Servers = append(lb.Servers[:i], lb.Servers[i+1:]...)
            delete(lb.serverInfo, server)
            break
        }
    }
}

func (lb *models.LoadBalancer) getNextServer() *models.Server {
    lb.lock.Lock()
    defer lb.lock.Unlock()

    var leastConnections *serverInfo
    for _, info := range lb.serverInfo {
        if leastConnections == nil || info.connections < leastConnections.connections {
            leastConnections = info
        }
    }

    if leastConnections == nil {
        return nil
    }

    leastConnections.connections++
    leastConnections.lastUsed = time.Now()

    return leastConnections.server
}
