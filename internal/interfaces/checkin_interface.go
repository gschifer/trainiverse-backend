package interfaces

import (
	"time"
	"trainiverse-backend/internal/models"
)

type CheckinInterface interface {
	HasCheckedInToday(userID string) (bool, error)
	SaveCheckinToDB(data models.CheckinData) error
	GetPathImage(userID string) (string, error)
	GetCheckinDate(userId string) (time.Time, error)
}
