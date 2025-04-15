package main

import (
	"log"
	"net/http"
	"os"
	"trainiverse-backend/handlers"
)

func main() {
	http.HandleFunc("/upload", handlers.UploadHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
