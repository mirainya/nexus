package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	config.C = &config.Config{
		Server: config.ServerConfig{JWTSecret: "test-secret-key-that-is-at-least-32-chars"},
	}
	logger.Init()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to open test db: " + err.Error())
	}
	testDB = db
	db.AutoMigrate(&model.APIKey{}, &model.APIUsage{})

	os.Exit(m.Run())
}

func seedAPIKey(t *testing.T, key string, active bool, expiresAt *time.Time) model.APIKey {
	t.Helper()
	ak := model.APIKey{Name: "test", Key: key, Active: active, ExpiresAt: expiresAt}
	if err := testDB.Create(&ak).Error; err != nil {
		t.Fatalf("seed api key: %v", err)
	}
	return ak
}

// --- APIKeyAuth Tests ---

func TestAPIKeyAuth_NoKey(t *testing.T) {
	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "nonexistent-key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_InactiveKey(t *testing.T) {
	ak := seedAPIKey(t, "inactive-key-001", true, nil)
	testDB.Model(&ak).Update("active", false)

	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "inactive-key-001")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_ExpiredKey(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	seedAPIKey(t, "expired-key-001", true, &past)

	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "expired-key-001")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_ValidKey_Header(t *testing.T) {
	ak := seedAPIKey(t, "valid-header-key-001", true, nil)

	var capturedID any
	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) {
		capturedID, _ = c.Get("api_key_id")
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "valid-header-key-001")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedID != ak.ID {
		t.Errorf("expected api_key_id %d, got %v", ak.ID, capturedID)
	}
}

func TestAPIKeyAuth_ValidKey_QueryParam(t *testing.T) {
	seedAPIKey(t, "valid-query-key-001", true, nil)

	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?api_key=valid-query-key-001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAPIKeyAuth_ValidKey_Bearer(t *testing.T) {
	seedAPIKey(t, "valid-bearer-key-001", true, nil)

	r := gin.New()
	r.GET("/test", APIKeyAuth(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-bearer-key-001")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- QuotaCheck Tests ---

func TestQuotaCheck_NoLimits(t *testing.T) {
	ak := seedAPIKey(t, "quota-nolimit-001", true, nil)

	r := gin.New()
	r.GET("/test", func(c *gin.Context) { c.Set("api_key_id", ak.ID); c.Next() }, QuotaCheck(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestQuotaCheck_DailyRequestExceeded(t *testing.T) {
	ak := model.APIKey{Name: "quota-daily-req", Key: "quota-daily-req-001", Active: true, DailyLimit: 5}
	testDB.Create(&ak)

	today := time.Now().Format("2006-01-02")
	testDB.Create(&model.APIUsage{APIKeyID: ak.ID, Date: today, Requests: 5, Tokens: 0})

	r := gin.New()
	r.GET("/test", func(c *gin.Context) { c.Set("api_key_id", ak.ID); c.Next() }, QuotaCheck(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

func TestQuotaCheck_DailyTokenExceeded(t *testing.T) {
	ak := model.APIKey{Name: "quota-daily-tok", Key: "quota-daily-tok-001", Active: true, DailyTokens: 1000}
	testDB.Create(&ak)

	today := time.Now().Format("2006-01-02")
	testDB.Create(&model.APIUsage{APIKeyID: ak.ID, Date: today, Requests: 1, Tokens: 1000})

	r := gin.New()
	r.GET("/test", func(c *gin.Context) { c.Set("api_key_id", ak.ID); c.Next() }, QuotaCheck(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

func TestQuotaCheck_MonthlyRequestExceeded(t *testing.T) {
	ak := model.APIKey{Name: "quota-monthly-req", Key: "quota-monthly-req-001", Active: true, MonthlyLimit: 100}
	testDB.Create(&ak)

	today := time.Now().Format("2006-01-02")
	testDB.Create(&model.APIUsage{APIKeyID: ak.ID, Date: today, Requests: 100, Tokens: 0})

	r := gin.New()
	r.GET("/test", func(c *gin.Context) { c.Set("api_key_id", ak.ID); c.Next() }, QuotaCheck(testDB), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

// --- JWTAuth Tests ---

func TestJWTAuth_NoToken(t *testing.T) {
	r := gin.New()
	r.GET("/test", JWTAuth(), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	r := gin.New()
	r.GET("/test", JWTAuth(), func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	token, err := GenerateToken(1, "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	var capturedUserID any
	r := gin.New()
	r.GET("/test", JWTAuth(), func(c *gin.Context) {
		capturedUserID, _ = c.Get("user_id")
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedUserID == nil {
		t.Error("expected user_id in context")
	}
}
