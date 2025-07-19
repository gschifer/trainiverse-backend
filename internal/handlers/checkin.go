package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"trainiverse-backend/internal/firebase"
	"trainiverse-backend/internal/interfaces"
	"trainiverse-backend/internal/models"

	"trainiverse-backend/internal/utils"
)

type CheckinService struct {
	checkinRepo    interfaces.CheckinInterface
	FirebaseClient firebase.FirebaseInterface
	ExifDecoder    utils.ExifDecoderInterface
	ImageSaver     utils.ImageSaverInterface
}

func NewCheckinService(checkinRepo interfaces.CheckinInterface, exifDecoder utils.ExifDecoderInterface) *CheckinService {
	return &CheckinService{
		checkinRepo:    checkinRepo,
		FirebaseClient: &firebase.FirebaseClient{},
		ExifDecoder:    exifDecoder,
		ImageSaver:     &utils.ImageSaver{},
	}
}

func (service CheckinService) CheckinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := service.FirebaseClient.GetUserID(r)
	if userID == "" {
		http.Error(w, `{"error":"userID not found"}`, http.StatusUnauthorized)
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

	x, err := service.ExifDecoder.Decode(file)
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

	hasChecked, err := service.checkinRepo.HasCheckedInToday(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to check in the database"}`, http.StatusInternalServerError)
		return
	}

	if hasChecked {
		http.Error(w, `{"error":"user has already checked in today"}`, http.StatusConflict)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, `{"error":"failed to reset file pointer"}`, http.StatusInternalServerError)
		return
	}

	err = service.ImageSaver.SaveImage(file, header.Filename, "checkins", userID)
	if err != nil {
		http.Error(w, `{"error":"failed to save image"}`, http.StatusInternalServerError)
		return
	}

	checkinInfo := models.CheckinData{
		UserID:      userID,
		CheckinDate: timestamp,
		FileName:    header.Filename,
	}

	if err := service.checkinRepo.SaveCheckinToDB(checkinInfo); err != nil {
		http.Error(w, `{"error":"failed to save checkin to database"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Check-in saved successfully",
		// "timestamp": timestamp,
	})
}
