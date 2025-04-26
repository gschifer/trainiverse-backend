package interfaces

import "trainiverse-backend/internal/models"

type CheckinInterface interface {
	HasCheckedInToday(userID string) (bool, error)
	SaveCheckinToDB(data models.CheckinData) error
}