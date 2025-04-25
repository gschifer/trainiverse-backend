package repository

import (
	"database/sql"
	db "trainiverse-backend/internal/db"
	"trainiverse-backend/internal/interfaces"
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
	err := db.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}


