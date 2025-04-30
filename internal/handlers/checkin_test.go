package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	fireBaseMocks "trainiverse-backend/internal/firebase/mocks"
	"trainiverse-backend/internal/interfaces/mocks"
	utilMock "trainiverse-backend/internal/utils/mocks"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckinHandler_Success(t *testing.T) {
	// Prepare a fake image file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.jpg")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake image content"))
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/checkin", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	// Create mock instances
	mockCheckinRepo := new(mocks.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockExifDecoder := new(utilMock.MockExifDecoderInterface)

	// Set expectations for mocks
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(false, nil)
	mockCheckinRepo.On("SaveCheckinToDB", mock.AnythingOfType("models.CheckinData")).Return(nil)

	file, _ := os.Open("testdata/test_image.jpeg")
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	// Initialize service with mocks
	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, string(body), "Check-in saved successfully")

	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}
