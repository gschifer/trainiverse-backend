package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"trainiverse-backend/internal/db"
	"trainiverse-backend/internal/firebase"
	"trainiverse-backend/internal/models"

	"trainiverse-backend/internal/utils"

	"github.com/rwcarlsen/goexif/exif"
)

// TODO this will be removed to use a database
var checkinTimes = make(map[string]time.Time)

func CheckinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	userID := firebase.GetUserID(r)
	if userID == "" {
		http.Error(w, `{"error":"userID not found"}`, http.StatusUnauthorized)
		return
	}

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

	println("EXIF data:", x)

	timestamp, err := utils.ExtractTime(x)
	if err != nil || !utils.IsToday(timestamp) {
		http.Error(w, `{"error":"photo must be from today"}`, http.StatusBadRequest)
		return
	}

	// Check if the user has already checked in today
	lastCheckInTime, exists := checkinTimes[userID]
	if exists && lastCheckInTime.YearDay() == time.Now().YearDay() {
		http.Error(w, "User has already checked in today", http.StatusConflict)
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

	checkinInfo := models.CheckinData{
		UserID:         userID,
		CheckinDate:    time.Now(),
		PhotoTimestamp: timestamp,
	}

	if err := saveCheckinToDB(checkinInfo); err != nil {
		http.Error(w, `{"error":"failed to save checkin to database"}`, http.StatusInternalServerError)
		return
	}

	checkinTimes[userID] = timestamp

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":   "Check-in saved successfully",
		"timestamp": timestamp,
	})
}

func saveCheckinToDB(data models.CheckinData) error {
	// Prepare the SQL query
	query := `
		INSERT INTO checkins (user_id, image_path, checkin_date)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	// Get the image path where the image was saved
	imagePath := fmt.Sprintf("checkins/%s_%s.jpg", data.UserID, time.Now().Format("20060102_150405"))

	var id int
	err := database.DB.QueryRow(query, data.UserID, imagePath, data.CheckinDate).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert checkin data into DB: %w", err)
	}

	fmt.Printf("Check-in saved with ID %d\n", id)
	return nil
}
