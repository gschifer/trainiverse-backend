package main

import (
	"log"
	"net/http"
	"os"
	database "trainiverse-backend/internal/db"
	"trainiverse-backend/internal/firebase"
	"trainiverse-backend/internal/handlers"
	"trainiverse-backend/internal/repository"
	"trainiverse-backend/internal/utils"
)

func main() {
	firebase.StartFirebase()
	database.InitDatabase()

	checkInRepo := repository.NewCheckinRepo(database.DB)
	checkoutRepo := repository.NewCheckoutRepo(database.DB)
	checkout := handlers.NewCheckoutService(checkInRepo, checkoutRepo)
	checkin := handlers.NewCheckinService(checkInRepo, &utils.ExifDecoder{})

	http.HandleFunc("/checkin", firebase.FirebaseAuthMiddleware(checkin.CheckinHandler))
	http.HandleFunc("/checkout", firebase.FirebaseAuthMiddleware(checkout.CheckoutHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Println(http.ListenAndServe(":"+port, nil))
}
