package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Allow CORS (useful for frontend testing)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse uploaded file (limit to 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing the file", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode EXIF metadata
	x, err := exif.Decode(file)
	if err != nil {
		http.Error(w, `{"error": "Failed to read EXIF metadata"}`, http.StatusInternalServerError)
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

func main() {
	http.HandleFunc("/upload", uploadHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default local
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
