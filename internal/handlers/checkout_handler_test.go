package handlers

import (
	"bytes"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/mock"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	fireBaseMocks "trainiverse-backend/internal/firebase/mocks"
	checkinMock "trainiverse-backend/internal/interfaces/mocks"
	checkoutMock "trainiverse-backend/internal/interfaces/mocks"
	utilMock "trainiverse-backend/internal/utils/mocks"

	"github.com/stretchr/testify/assert"
)

func setupCheckoutMocks(t *testing.T) (*http.Request, *httptest.ResponseRecorder, *checkinMock.MockCheckinInterface, *fireBaseMocks.MockFirebaseInterface, *utilMock.MockExifDecoderInterface, *utilMock.MockImageSaverInterface, *checkoutMock.MockCheckoutInterface) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.jpg")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake image content"))
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/checkout", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockExifDecoder := new(utilMock.MockExifDecoderInterface)
	mockImageSaver := new(utilMock.MockImageSaverInterface)
	mockCheckoutRepo := new(checkoutMock.MockCheckoutInterface)
	return req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, mockCheckoutRepo
}

func TestCheckoutHandler_MethodNotAllowed(t *testing.T) {
	_, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	req := httptest.NewRequest(http.MethodGet, "/checkout", nil)
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Contains(t, string(body), "Method not allowed")
}

func TestCheckoutHandler_Unauthenticated(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("")
	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "userID not found")
	mockFirebaseClient.AssertExpectations(t)
}

func TestCheckoutHandler_FileMissing(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/checkout", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockImageSaver := new(utilMock.MockImageSaverInterface)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ImageSaver:      mockImageSaver,
	}

	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "Error retrieving the file")
	mockFirebaseClient.AssertExpectations(t)
}

func TestCheckoutHandler_ParseMultipartFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBuffer([]byte("invalid body")))
	req.Header.Set("Content-Type", "multipart/form-data")
	rec := httptest.NewRecorder()

	mockCheckinRepo := new(checkinMock.MockCheckinInterface)
	mockFirebaseClient := new(fireBaseMocks.MockFirebaseInterface)
	mockImageSaver := new(utilMock.MockImageSaverInterface)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ImageSaver:      mockImageSaver,
	}

	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "Error parsing the file")
	mockFirebaseClient.AssertExpectations(t)
}

func TestCheckoutHandler_UserNotFound(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("")
	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "userID not found")
	mockFirebaseClient.AssertExpectations(t)
}

func TestCheckoutHandler_SaveImageError(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(-40*time.Minute), nil)
	mockImageSaver.On("SaveImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))

	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)

	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "failed to save image")
	mockImageSaver.AssertExpectations(t)
}

func TestCheckoutHandler_EXIFError(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(-40*time.Minute), nil)
	mockExifDecoder.On("Decode", mock.Anything).Return(nil, assert.AnError)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "could not read EXIF")
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckoutHandler_PhotoNotFromToday(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(-40*time.Minute), nil)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Add(-24*time.Hour).Format(time.DateTime))
	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)
	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "photo must be from today")
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckoutHandler_CheckinNotFound(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(false, nil)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))
	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)
	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "user has not checked in today")
	mockCheckinRepo.AssertExpectations(t)
}

func TestCheckoutHandler_CheckoutBeforeCheckin(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(10*time.Minute), nil)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))
	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)
	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "checkout time must be after checkin time")
	mockCheckinRepo.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckoutHandler_CheckoutLessThan20Minutes(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, _ := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(-10*time.Minute), nil)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))
	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)
	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "must wait at least 20 minutes before checkout")
	mockCheckinRepo.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
}

func TestCheckoutHandler_Success(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, mockCheckoutRepo := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(-40*time.Minute), nil)
	mockImageSaver.On("SaveImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCheckoutRepo.On("SaveCheckout", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))
	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)
	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		checkoutRepo:    mockCheckoutRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, string(body), "Check-out saved successfully")
	mockCheckinRepo.AssertExpectations(t)
	mockFirebaseClient.AssertExpectations(t)
	mockExifDecoder.AssertExpectations(t)
	mockImageSaver.AssertExpectations(t)
	mockCheckoutRepo.AssertExpectations(t)
}

func TestCheckoutHandler_FailedToSaveCheckoutToDatabase(t *testing.T) {
	req, rec, mockCheckinRepo, mockFirebaseClient, mockExifDecoder, mockImageSaver, mockCheckoutRepo := setupCheckoutMocks(t)
	mockFirebaseClient.On("GetUserID", req).Return("test-user-id")
	mockCheckinRepo.On("HasCheckedInToday", "test-user-id").Return(true, nil)
	mockCheckinRepo.On("GetCheckinDate", "test-user-id").Return(time.Now().Add(-40*time.Minute), nil)
	mockImageSaver.On("SaveImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCheckoutRepo.On("SaveCheckout", mock.Anything).Return(assert.AnError)

	imagePath := "../test_data/test_image.jpeg"
	setExifDateTime(t, imagePath, time.Now().Format(time.DateTime))
	file, _ := os.Open(imagePath)
	x, _ := exif.Decode(file)
	mockExifDecoder.On("Decode", mock.Anything).Return(x, nil)

	service := &CheckoutService{
		checkinRepo:     mockCheckinRepo,
		checkoutRepo:    mockCheckoutRepo,
		firebaseService: mockFirebaseClient,
		ExifDecoder:     mockExifDecoder,
		ImageSaver:      mockImageSaver,
	}
	service.CheckoutHandler(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "failed to save checkout to database")
	mockCheckoutRepo.AssertExpectations(t)
}
