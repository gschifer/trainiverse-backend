package utils

import (
	"fmt"
	"time"
)

const ImagePathPattern = "checkins/%s_%s.jpg"

func BuildImagePath(userID string) string {
    timestamp := time.Now().Format("20060102_150405")
    return fmt.Sprintf(ImagePathPattern, userID, timestamp)
}