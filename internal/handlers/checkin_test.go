package handlers

import (
	"bytes"
	"errors"

	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"
	fireBaseMocks "trainiverse-backend/internal/firebase/mocks"
	checkinMock "trainiverse-backend/internal/interfaces/mocks"
	utilMock "trainiverse-backend/internal/utils/mocks"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckinHandler_When_PhotoIsNot_FromToday(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockExifDecoder.On("Decode", mock.Anything).Return(&exif.Exif{}, nil)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id") // Simulate a valid user ID

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"photo must be from today"}`)

	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}
func TestCheckinHandler_When_File_Is_Missing(t *testing.T) {

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.Close() // Close the writer to finalize the multipart form
	req := httptest.NewRequest(http.MethodPost, "/checkin", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockExifDecoder := new(utilMock.MockExifDecoderInterface)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id") // Simulate a valid user ID

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"file is required"}`)

	mockFirebaseClient.AssertExpectations(t)
}
func TestCheckinHandler_ParseMultipartFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/checkin", bytes.NewBuffer([]byte("invalid body")))
	req.Header.Set("Content-Type", "multipart/form-data")
	rec := httptest.NewRecorder()

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockExifDecoder := new(utilMock.MockExifDecoderInterface)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id") // Simulate a valid user ID

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"failed to parse file"}`)

	mockFirebaseClient.AssertExpectations(t)
}
func TestCheckinHandler_When_Method_Is_NotAllowed(t *testing.T) {
	// Prepare a fake request with an invalid HTTP method (e.g., GET)
	req := httptest.NewRequest(http.MethodGet, "/checkin", nil)
	rec := httptest.NewRecorder()

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockExifDecoder := new(utilMock.MockExifDecoderInterface)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Contains(t, string(body), "Method not allowed")
}
func TestCheckinHandler_When_UserNotFound(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("") // Simulate user not found

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"userID not found"}`)

	mockFirebaseClient.AssertExpectations(t)
}

func TestCheckinHandler_When_Photo_Not_From_Today(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")

	imagePath := "../testdata/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Add(-24*time.Hour).Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"photo must be from today"}`)

	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func setExifDateTime(t *testing.T, filepath string, datetime string) {
	t.Helper()
	cmd := exec.Command("exiftool", "-overwrite_original", "-DateTimeOriginal="+datetime, filepath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set EXIF DateTimeOriginal: %v", err)
	}
}

func TestCheckinHandler_Error_In_The_Database_When_Try_To_Check_If_User_Already_Checkin(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(false, errors.New("failed to check in the database"))

	imagePath := "../testdata/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Expect 409 Conflict
	assert.Contains(t, string(body), `{"error":"failed to check in the database"}`)

	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckinHandler_UserAlreadyCheckedIn(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)

	imagePath := "../testdata/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusConflict, resp.StatusCode) // Expect 409 Conflict
	assert.Contains(t, string(body), `{"error":"user has already checked in today"}`)

	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckinHandler_Error_On_SaveImage_In_The_Database(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(false, nil)
	mockCheckinRepo.On("SaveCheckinToDB", mock.AnythingOfType("models.CheckinData")).Return(errors.New("failed to save checkin to database"))
	mockImageSaver.On("SaveImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	imagePath := "../testdata/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"failed to save checkin to database"}`)

	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckinHandler_Success(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(false, nil)
	mockCheckinRepo.On("SaveCheckinToDB", mock.AnythingOfType("models.CheckinData")).Return(nil)
	mockImageSaver.On("SaveImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	imagePath := "../testdata/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
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

func TestCheckinHandler_FailedToSaveImage(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver := setupMocks(t)

	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(false, nil)
	mockImageSaver.On("SaveImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to save image"))

	imagePath := "../testdata/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckinService{
		checkinRepo:    mockCheckinRepo,
		FirebaseClient: mockFirebaseClient,
		ExifDecoder:    mockExifDecoder,
		ImageSaver:     mockImageSaver,
	}

	service.CheckinHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Expect 500 Internal Server Error
	assert.Contains(t, string(body), `{"error":"failed to save image"}`)

	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func setupMocks(t *testing.T) (*http.Request, *httptest.ResponseRecorder, *checkinMock.MockCheckinInterface, *fireBaseMocks.MockFirebaseInterface, *utilMock.MockExifDecoderInterface, *utilMock.MockImageSaverInterface) {
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

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockExifDecoder := new(utilMock.MockExifDecoderInterface)
	mockImageSaver := new(utilMock.MockImageSaverInterface)
	return req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver
}
