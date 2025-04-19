package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func ExtractTime(x *exif.Exif) (time.Time, error) {
	return x.DateTime()
}

func IsToday(t time.Time) bool {
	now := time.Now()
	return now.Year() == t.Year() && now.YearDay() == t.YearDay()
}

func SaveImage(file multipart.File, originalName, folder, userID string) error {
	timestamp := time.Now().Format(time.RFC3339)
	filename := fmt.Sprintf("%s_%s%s", userID, timestamp, filepath.Ext(originalName))
	outPath := filepath.Join("storage", folder, filename)

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	dst, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(dst, file)
	return err
}
