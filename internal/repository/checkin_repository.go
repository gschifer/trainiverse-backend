package repository

import (
	"database/sql"
	"fmt"
	"time"
	"trainiverse-backend/internal/interfaces"
	"trainiverse-backend/internal/models"
	"trainiverse-backend/internal/utils"
)

type CheckinRepository struct {
	DB *sql.DB
}

func NewCheckinRepo(db *sql.DB) interfaces.CheckinInterface {
	return &CheckinRepository{
		DB: db,
	}
}

func (checkinRepo *CheckinRepository) GetPathImage(userId string) (string, error) {
	query := `
		SELECT image_path 
		FROM checkins 
		WHERE user_id = $1 
		  AND DATE(checkin_date) = CURRENT_DATE;
	`

	var imagePath string
	err := checkinRepo.DB.QueryRow(query, userId).Scan(&imagePath)
	if err != nil {
		fmt.Println("Error fetching image path:", err)
	}

	return imagePath, nil
}

func (checkinRepo *CheckinRepository) GetCheckinDate(userId string) (time.Time, error) {
	query := `
		SELECT checkin_date 
		FROM checkins 
		WHERE user_id = $1 
		  AND DATE(checkin_date) = CURRENT_DATE;
	`

	var checkinDate time.Time
	err := checkinRepo.DB.QueryRow(query, userId).Scan(&checkinDate)
	if err != nil {
		fmt.Println("Error fetching checkin date:", err)
	}

	return checkinDate, nil
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
	imagePath := utils.BuildImagePathForCheckin(data)
	//imagePath := utils.BuildImagePathForCheckin(data)

	var id int
	err := checkinRepo.DB.QueryRow(query, data.UserID, imagePath, data.CheckinDate).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert checkin data into DB: %w", err)
	}

	fmt.Printf("Check-in saved with ID %d\n", id)
	return nil
}
