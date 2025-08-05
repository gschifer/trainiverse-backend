package repository

import (
	"database/sql"
	"fmt"
	"trainiverse-backend/internal/interfaces"
	"trainiverse-backend/models"

	"trainiverse-backend/internal/utils"
)

type CheckoutRepository struct {
	DB *sql.DB
}

func NewCheckoutRepo(db *sql.DB) interfaces.CheckoutInterface {
	return &CheckoutRepository{
		DB: db,
	}
}

func (checkoutRepo *CheckoutRepository) SaveCheckout(checkoutData models.CheckoutData) error {
	query := `
		INSERT INTO checkouts (user_id, image_path, checkout_date)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	imagePath := utils.BuildImagePathForCheckout(checkoutData)

	var id int
	err := checkoutRepo.DB.QueryRow(query, checkoutData.UserID, imagePath, checkoutData.CheckoutDate).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert checkin data into DB: %w", err)
	}

	fmt.Printf("Checkout saved with ID %d\n", id)
	return nil
}
