package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"trainiverse-backend/internal/firebase"
	"trainiverse-backend/internal/interfaces"

	"trainiverse-backend/internal/utils"

	"github.com/rwcarlsen/goexif/exif"
)

type CheckoutService struct {
	checkinRepo interfaces.CheckinInterface
	firebaseService firebase.FirebaseInterface
}

func NewCheckoutService(checkinRepo interfaces.CheckinInterface) *CheckoutService {
	return &CheckoutService{
		checkinRepo: checkinRepo,
		firebaseService: &firebase.FirebaseClient{},
	}
}

func (service *CheckoutService) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		utils.HandleError(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	println("EXIF data:", x)

	timestamp, err := utils.ExtractTime(x)
	if err != nil || !utils.IsToday(timestamp) {
		http.Error(w, `{"error":"photo must be from today"}`, http.StatusBadRequest)
		return
	}

	userID := service.firebaseService.GetUserID(r)
	ok, err := service.checkinRepo.HasCheckedInToday(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to check in the database"}`, http.StatusInternalServerError)
	}
	if !ok {
		http.Error(w, `{"error":"user has not checked in today"}`, http.StatusBadRequest)
	}

	// TODO Check in the DB the path of the check-in for the user to compare the times
	data, _ := LoadCheckinLog(filepath.Join("storage/checkins/%s.json"))
	parsedTime, _ := time.Parse(time.RFC3339, data.PhotoMetadataTime)

	if timestamp.Sub(parsedTime) < 20*time.Minute {
		http.Error(w, `{"error":"must wait at least 20 minutes before checkout"}`, http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, `{"error":"failed to reset file pointer"}`, http.StatusInternalServerError)
		return
	}
	err = utils.SaveImage(file, header.Filename, "checkouts", userID)
	if err != nil {
		http.Error(w, `{"error":"failed to save image"}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":     "Check-out saved successfully",
		// "checkin_at":  start,
		"checkout_at": timestamp,
		// "duration":    timestamp.Sub(start).String(),
	})
}

type CheckinLog struct {
	UserID            string `json:"userId"`
	CheckinDate       string `json:"checkinDate"`
	PhotoMetadataTime string `json:"photoTimestamp"`
}

// LoadCheckinLog loads and parses a check-in log JSON file from disk.
func LoadCheckinLog(filePath string) (*CheckinLog, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var log CheckinLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &log, nil
}
