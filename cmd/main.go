package main

import (
	"log"
	"net/http"
	"os"
	database "trainiverse-backend/internal/db"
	"trainiverse-backend/internal/firebase"
	"trainiverse-backend/internal/handlers"
	"trainiverse-backend/internal/repository"
)

func main() {
	firebase.StartFirebase()

	database.InitDatabase()

	checkInRepo := repository.NewCheckinRepo(database.DB)
	checkInService := handlers.NewCheckoutService(checkInRepo)

	http.HandleFunc("/checkin", firebase.FirebaseAuthMiddleware(handlers.CheckinHandler))
	http.HandleFunc("/checkout", firebase.FirebaseAuthMiddleware(checkInService.CheckoutHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Println(http.ListenAndServe(":"+port, nil))
}
