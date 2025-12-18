package router

import (
	"bytes"
	"encoding/json"
	"mygram-api/database"
	"mygram-api/helpers"
	"mygram-api/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TokenMock adalah mock untuk CreateToken menggunakan testify/mock
type TokenMock struct {
	mock.Mock
}

func (m *TokenMock) CreateToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func setupInMemoryDB(t *testing.T) *gorm.DB {
	// in-memory sqlite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
}

func TestLogin_Success(t *testing.T) {
	// set JWT_SECRET_KEY for jwt signing (CreateToken default uses this, but we mock CreateTokenFunc)
	os.Setenv("JWT_SECRET_KEY", "testsecret")

	// setup db and insert user
	testDB := setupInMemoryDB(t)

	// create password hash
	hashed, err := helpers.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}
	user := models.User{
		Email:    "user@example.com",
		Password: hashed,
	}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// override database.GetDB to return our in-memory DB
	database.GetDB = func() *gorm.DB {
		return testDB
	}

	// mock CreateTokenFunc using testify/mock
	tokenMock := &TokenMock{}
	// we don't need exact userID here, so allow mock.Anything for userID
	tokenMock.On("CreateToken", mock.Anything, "user@example.com").Return("mocked-token", nil)

	// override helpers.CreateTokenFunc to the mock's method
	helpers.CreateTokenFunc = func(userID uuid.UUID, email string) (string, error) {
		return tokenMock.CreateToken(userID, email)
	}

	// init router and call endpoint
	router := SetupRouter()

	reqBody := map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// check response: top-level token
	tokenVal, ok := resp["token"].(string)
	assert.True(t, ok, "token should be a string in response")
	assert.Equal(t, "mocked-token", tokenVal)

	// assert mock called
	tokenMock.AssertCalled(t, "CreateToken", mock.Anything, "user@example.com")
}

func TestLogin_InvalidPassword(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "testsecret")

	// setup db and insert user
	testDB := setupInMemoryDB(t)

	hashed, err := helpers.HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}
	user := models.User{
		Email:    "user2@example.com",
		Password: hashed,
	}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// override database.GetDB
	database.GetDB = func() *gorm.DB {
		return testDB
	}

	// Use real CreateTokenFunc (no need to mock) — but it should not be called because password is invalid.

	router := SetupRouter()

	reqBody := map[string]string{
		"email":    "user2@example.com",
		"password": "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Invalid email or password", resp["message"])
}

func TestSetupRouter_SwaggerEndpointExists(t *testing.T) {
	t.Parallel()

	r := SetupRouter()
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// We just assert that the route isn't 404. The swagger handler serves static UI.
	assert.NotEqual(t, http.StatusNotFound, w.Code, "swagger endpoint should be registered")
}
