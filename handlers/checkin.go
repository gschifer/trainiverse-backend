package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
	"trainiverse-backend/utils"

	"github.com/rwcarlsen/goexif/exif"
)

var checkinTimes = make(map[string]time.Time)

func CheckinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, `{"error":"failed to parse file"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			http.Error(w, `{"error":"failed to close file"}`, http.StatusInternalServerError)
		}
	}()

	x, err := exif.Decode(file)
	if err != nil {
		http.Error(w, `{"error":"could not read EXIF"}`, http.StatusInternalServerError)
		return
	}

	userID := r.Header.Get("UserID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check if the user has already checked in today
	lastCheckInTime, exists := checkinTimes[userID]
	if exists && lastCheckInTime.YearDay() == time.Now().YearDay() {
		http.Error(w, "User has already checked in today", http.StatusConflict)
		return
	}

	timestamp, err := utils.ExtractTime(x)
	if err != nil || !utils.IsToday(timestamp) {
		http.Error(w, `{"error":"photo must be from today"}`, http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, `{"error":"failed to reset file pointer"}`, http.StatusInternalServerError)
		return
	}

	err = utils.SaveImage(file, header.Filename, "checkins", userID)
	if err != nil {
		http.Error(w, `{"error":"failed to save image"}`, http.StatusInternalServerError)
		return
	}

	checkinTimes[userID] = timestamp

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":   "Check-in saved successfully",
		"timestamp": timestamp,
	})
}
