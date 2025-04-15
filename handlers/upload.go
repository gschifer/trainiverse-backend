package handlers

import (
	"encoding/json"
	"net/http"
	"trainiverse-backend/utils"

	"github.com/rwcarlsen/goexif/exif"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
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
	file, _, err := r.FormFile("file")
	if err != nil {
		utils.HandleError(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode EXIF metadata
	x, err := exif.Decode(file)
	if err != nil {
		utils.HandleError(w, "Failed to read EXIF metadata", http.StatusInternalServerError)
		return
	}

	// Try to extract GPS coordinates
	lat, long, _ := x.LatLong()

	// Prepare response structure
	response := map[string]any{
		"latitude":  lat,
		"longitude": long,
	}

	// Add EXIF metadata (as JSON)
	if data, err := x.MarshalJSON(); err == nil {
		var exifData map[string]any
		if json.Unmarshal(data, &exifData) == nil {
			response["exif"] = exifData
		}
	}

	// Return JSON response
	json.NewEncoder(w).Encode(response)
}
