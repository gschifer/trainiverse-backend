package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type ImageSaverInterface interface {
	SaveImage(file multipart.File, originalName, folder, userID string) error
}

type ImageSaver struct{}

func (service *ImageSaver) SaveImage(file multipart.File, originalName, folder, userID string) error {
	timestamp := time.Now().Format("20060102_150405")
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
