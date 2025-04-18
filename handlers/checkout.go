package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
	"trainiverse-backend/utils"

	"github.com/rwcarlsen/goexif/exif"
)

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
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

	timestamp, err := utils.ExtractTime(x)
	if err != nil || !utils.IsToday(timestamp) {
		http.Error(w, `{"error":"photo must be from today"}`, http.StatusBadRequest)
		return
	}

	userID := "123"
	start, ok := checkinTimes[userID]
	if !ok {
		http.Error(w, `{"error":"no check-in found"}`, http.StatusBadRequest)
		return
	}

	if timestamp.Sub(start) < 20*time.Minute {
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
		"checkin_at":  start,
		"checkout_at": timestamp,
		"duration":    timestamp.Sub(start).String(),
	})
}
