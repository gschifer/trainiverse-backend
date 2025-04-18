package main

import (
	"log"
	"net/http"
	"os"
	"trainiverse-backend/handlers"
)

func main() {
	http.HandleFunc("/checkin", handlers.CheckinHandler)
	http.HandleFunc("/checkout", handlers.CheckinHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
