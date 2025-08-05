package models

import "time"

type CheckoutData struct {
	UserID       string    `json:"userID"`
	CheckoutDate time.Time `json:"checkinDate"`
	FileName     string    `json:"fileName"`
}
