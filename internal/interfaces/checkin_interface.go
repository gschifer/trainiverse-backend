package interfaces

type CheckinInterface interface {
	HasCheckedInToday(userID string) (bool, error)
	// SaveCheckin(userID string, checkinDate string, photoTimestamp string) error
}