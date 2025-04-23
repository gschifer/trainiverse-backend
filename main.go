package main

import (
	"log"
	"net/http"
	"os"
	"trainiverse-backend/database"
	"trainiverse-backend/firebase"
	"trainiverse-backend/handlers"
)

func main() {
	firebase.StartFirebase()

	database.InitDatabase()

	http.HandleFunc("/checkin", firebase.FirebaseAuthMiddleware(handlers.CheckinHandler))
	http.HandleFunc("/checkout", firebase.FirebaseAuthMiddleware(handlers.CheckoutHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
