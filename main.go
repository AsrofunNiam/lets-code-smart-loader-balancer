package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ServerStats holds the details of a backend server.
type ServerStats struct {
	Address string
	Load    float64
}

var servers = []ServerStats{
	{"http://localhost:8081", 0.0},
	{"http://localhost:8082", 0.0},
	{"http://localhost:8083", 0.0},
	{"http://localhost:8084", 0.0},
}

var mu sync.Mutex
var rng = rand.New(rand.NewSource(time.Now().UnixNano())) // Local random generator

// AIWeightedRouting selects the best server based on simulated AI scoring.
func AIWeightedRouting() string {
	mu.Lock()
	defer mu.Unlock()

	// Simulate load updates for each server
	for i := range servers {
		if i == 0 { // Simulate high load for the first server
			servers[i].Load = 0.9 + rng.Float64()*0.1 // Always high
			fmt.Println(servers[i].Load)
		} else {
			servers[i].Load = rng.Float64() // Normal random load
		}
	}

	// Choose the server with the lowest load
	bestServer := servers[0]
	for _, server := range servers {
		if server.Load < bestServer.Load {
			bestServer = server
		}
	}

	log.Printf("Routing to server: %s (Load: %.2f)", bestServer.Address, bestServer.Load)
	return bestServer.Address
}

func startBackendServer(address string, idx int) {
	mux := http.NewServeMux()

	//  Simulate server busy
	mux.HandleFunc("/simulate-load", func(w http.ResponseWriter, r *http.Request) {
		// Simulate heavy processing
		time.Sleep(5 * time.Second)
		fmt.Fprintf(w, "Simulating heavy load on server %d", idx+1)
		fmt.Println("Simulating heavy load on server", idx+1)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend Server %d responding", idx+1)
	})
	server := &http.Server{
		Addr:    address[len("http://"):],
		Handler: mux,
	}
	log.Printf("Starting backend server %d at %s", idx+1, address)
	log.Fatal(server.ListenAndServe())
}

func proxyHandler(c *gin.Context) {
	backendURL := AIWeightedRouting()
	resp, err := http.Get(backendURL + c.Request.URL.Path)
	if err != nil {
		log.Printf("Error forwarding request: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to forward request"})
		return
	}
	defer resp.Body.Close()

	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

func main() {
	// Start backend servers to simulate load
	for i, server := range servers {
		go startBackendServer(server.Address, i)
	}

	// Create API Gateway
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to AI-Powered API Gateway"})
	})
	r.Any("/proxy/*path", proxyHandler)

	log.Println("Starting API Gateway on port 8080")
	log.Fatal(r.Run(":8080"))
}
