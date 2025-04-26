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
	checkout := handlers.NewCheckoutService(checkInRepo)
	checkin := handlers.NewCheckinService(checkInRepo)

	http.HandleFunc("/checkin", firebase.FirebaseAuthMiddleware(checkin.CheckinHandler))
	http.HandleFunc("/checkout", firebase.FirebaseAuthMiddleware(checkout.CheckoutHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Println(http.ListenAndServe(":"+port, nil))
}
