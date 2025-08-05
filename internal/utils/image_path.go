package utils

import (
	"fmt"
	"path/filepath"
	"time"
	"trainiverse-backend/models"
)

const PathPatternForCheckin = "checkins/%s_%s%s"
const PathPatternForCheckout = "checkouts/%s_%s%s"

func BuildImagePathForCheckin(checkinData models.CheckinData) string {
	timestamp := time.Now().Format("20060102_150405")
	extensionFile := filepath.Ext(checkinData.FileName)
	return fmt.Sprintf(PathPatternForCheckin, checkinData.UserID, timestamp, extensionFile)
}

func BuildImagePathForCheckout(checkoutData models.CheckoutData) string {
	timestamp := time.Now().Format("20060102_150405")
	extensionFile := filepath.Ext(checkoutData.FileName)
	return fmt.Sprintf(PathPatternForCheckout, checkoutData.UserID, timestamp, extensionFile)
}
