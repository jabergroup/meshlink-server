package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"meshlink-server/internal/api"
	"meshlink-server/internal/store"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	// Railway sets PORT env variable
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}

	fmt.Println("============================================")
	fmt.Println("  MeshLink Signaling Server")
	fmt.Println("  Direct. Private. No Third Party.")
	fmt.Println("============================================")

	// Initialize store
	s := store.New()

	// Create router
	router := api.NewRouter(s)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("\n  Listening on http://0.0.0.0%s\n", addr)
	fmt.Printf("  API:        http://localhost%s/api/health\n", addr)
	fmt.Printf("  Dashboard:  http://localhost%s\n\n", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
