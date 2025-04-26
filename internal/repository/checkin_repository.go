package repository

import (
	"database/sql"
	"fmt"
	"time"
	"trainiverse-backend/internal/interfaces"
	"trainiverse-backend/internal/models"
)

type CheckinRepository struct {
	DB *sql.DB
}

func NewCheckinRepo(db *sql.DB) interfaces.CheckinInterface {
	return &CheckinRepository{
		DB: db,
	}
}

func (checkinRepo *CheckinRepository) HasCheckedInToday(userID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM checkins 
		WHERE user_id = $1 
		  AND DATE(checkin_date) = CURRENT_DATE;
	`
	var count int
	err := checkinRepo.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (checkinRepo *CheckinRepository) SaveCheckinToDB(data models.CheckinData) error {
	query := `
		INSERT INTO checkins (user_id, image_path, checkin_date)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	// Get the image path where the image was saved
	imagePath := fmt.Sprintf("checkins/%s_%s.jpg", data.UserID, time.Now().Format("20060102_150405"))

	var id int
	err := checkinRepo.DB.QueryRow(query, data.UserID, imagePath, data.CheckinDate).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert checkin data into DB: %w", err)
	}

	fmt.Printf("Check-in saved with ID %d\n", id)
	return nil
}
