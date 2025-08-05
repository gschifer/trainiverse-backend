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
	checkout := handlers.NewCheckoutService(checkInRepo, checkoutRepo, &utils.ExifDecoder{})
	checkin := handlers.NewCheckinService(checkInRepo, &utils.ExifDecoder{})

	http.HandleFunc("/checkin", firebase.AuthMiddleware(checkin.CheckinHandler))
	http.HandleFunc("/checkout", firebase.AuthMiddleware(checkout.CheckoutHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Println(http.ListenAndServe(":"+port, nil))
}
