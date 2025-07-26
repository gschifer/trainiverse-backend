package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestHasCheckedInToday(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer dbMock.Close()

	checkinRepo := NewCheckinRepo(dbMock)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM checkins WHERE user_id = \$1 AND DATE\(checkin_date\) = CURRENT_DATE`).
		WithArgs("userID").
		WillReturnRows(rows)

	checkedIn, err := checkinRepo.HasCheckedInToday("userID")
	assert.Nil(t, err, "Expected no error when checking check-in")
	assert.True(t, checkedIn, "Expected user to have checked in today")
}
