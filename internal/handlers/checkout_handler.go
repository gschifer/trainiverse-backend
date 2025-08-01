package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
	"trainiverse-backend/internal/firebase"
	"trainiverse-backend/internal/interfaces"
	"trainiverse-backend/models"

	"trainiverse-backend/internal/utils"

	"github.com/rwcarlsen/goexif/exif"
)

type CheckoutService struct {
	checkinRepo     interfaces.CheckinInterface
	checkoutRepo    interfaces.CheckoutInterface
	firebaseService firebase.Interface
	ImageSaver      utils.ImageSaverInterface
}

func NewCheckoutService(checkinRepo interfaces.CheckinInterface,
	checkoutRepo interfaces.CheckoutInterface) *CheckoutService {
	return &CheckoutService{
		checkinRepo:     checkinRepo,
		checkoutRepo:    checkoutRepo,
		firebaseService: &firebase.Client{},
		ImageSaver:      &utils.ImageSaver{},
	}
}

func (service *CheckoutService) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		utils.HandleError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := service.firebaseService.GetUserID(r)
	if userID == "" {
		http.Error(w, `{"error":"userID not found"}`, http.StatusUnauthorized)
		return
	}

	// Parse uploaded file (limit to 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		utils.HandleError(w, "Error parsing the file", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.HandleError(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			utils.HandleError(w, "Failed to close the file", http.StatusInternalServerError)
		}
	}()

	x, err := exif.Decode(file)
	if err != nil {
		http.Error(w, `{"error":"could not read EXIF"}`, http.StatusInternalServerError)
		return
	}

	checkoutDateTime, err := utils.ExtractTime(x)
	if err != nil || !utils.IsToday(checkoutDateTime) {
		http.Error(w, `{"error":"photo must be from today"}`, http.StatusBadRequest)
		return
	}

	ok, err := service.checkinRepo.HasCheckedInToday(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to check in the database"}`, http.StatusInternalServerError)
	}
	if !ok {
		http.Error(w, `{"error":"user has not checked in today"}`, http.StatusBadRequest)
	}

	checkinDateTime, _ := service.checkinRepo.GetCheckinDate(userID)

	if !checkoutDateTime.After(checkinDateTime) {
		http.Error(w, `{"error":"checkout time must be after checkin time"}`, http.StatusBadRequest)
		return
	}

	if checkoutDateTime.Sub(checkinDateTime) < 20*time.Minute {
		http.Error(w, `{"error":"must wait at least 20 minutes before checkout"}`, http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, `{"error":"failed to reset file pointer"}`, http.StatusInternalServerError)
		return
	}
	err = service.ImageSaver.SaveImage(file, header.Filename, "checkouts", userID)
	if err != nil {
		http.Error(w, `{"error":"failed to save image"}`, http.StatusInternalServerError)
		return
	}

	checkout := models.CheckoutData{
		UserID:       userID,
		FileName:     header.Filename,
		CheckoutDate: checkoutDateTime,
	}

	if err := service.checkoutRepo.SaveCheckout(checkout); err != nil {
		http.Error(w, `{"error":"failed to save checkout to database"}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Check-out saved successfully",
		// "checkin_at":  start,
		"checkout_at": checkoutDateTime,
		// "duration":    checkoutDateTime.Sub(start).String(),
	})
}
