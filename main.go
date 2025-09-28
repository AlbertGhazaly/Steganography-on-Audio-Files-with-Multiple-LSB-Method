package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/handlers"
	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/middleware"
)

func main() {
	os.MkdirAll("./temp", 0755)

	http.HandleFunc("/api/health", middleware.CorsMiddleware(handlers.HealthCheck))
	http.HandleFunc("/api/embed", middleware.CorsMiddleware(handlers.EmbedHandler))
	http.HandleFunc("/api/extract", middleware.CorsMiddleware(handlers.ExtractHandler))
	http.HandleFunc("/api/capacity", middleware.CorsMiddleware(handlers.CapacityHandler))

	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fs)

	fmt.Println("Steganography Server starting on :8080")
	fmt.Println("API endpoints:")
	fmt.Println("  GET    /api/health")
	fmt.Println("  POST   /api/embed    - Embed secret file into MP3")
	fmt.Println("  POST   /api/extract  - Extract secret file from MP3")
	fmt.Println("  POST   /api/capacity - Calculate MP3 embedding capacity")
	fmt.Println("Frontend available at: http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
