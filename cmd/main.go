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
	apiKey := flag.String("api-key", "", "API key for authentication (or set MESHLINK_API_KEY env)")
	flag.Parse()

	// Railway sets PORT env variable
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}

	// Set API key
	if *apiKey != "" {
		api.SetAPIKey(*apiKey)
	}

	fmt.Println("============================================")
	fmt.Println("  MeshLink Signaling Server")
	fmt.Println("  Direct. Private. No Third Party.")
	fmt.Println("============================================")

	if key := os.Getenv("MESHLINK_API_KEY"); key != "" {
		fmt.Printf("\n  API Key:    %s...%s (from env)\n", key[:4], key[len(key)-4:])
	} else if *apiKey != "" {
		fmt.Printf("\n  API Key:    %s...%s (from flag)\n", (*apiKey)[:4], (*apiKey)[len(*apiKey)-4:])
	} else {
		fmt.Println("\n  WARNING: No API key set! Server is unprotected.")
		fmt.Println("  Set MESHLINK_API_KEY env or use -api-key flag.")
	}

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
