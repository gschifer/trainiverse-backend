package models

import "time"

type CheckinData struct {
	UserID      string    `json:"userID"`
	CheckinDate time.Time `json:"checkinDate"`
}
