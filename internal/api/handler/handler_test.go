package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/middleware"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testRouter *gin.Engine
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
	db.AutoMigrate(
		&model.User{},
		&model.APIKey{},
		&model.PromptTemplate{},
		&model.Pipeline{},
		&model.PipelineStep{},
		&model.Document{},
		&model.Entity{},
		&model.Relation{},
		&model.Job{},
		&model.JobStepLog{},
		&model.Review{},
		&model.LLMProvider{},
		&model.Credential{},
		&model.APIUsage{},
	)
	model.SetDB(db)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	db.Create(&model.User{Username: "admin", Password: string(hashed), Role: "admin"})

	testRouter = setupTestRouter()

	os.Exit(m.Run())
}

func setupTestRouter() *gin.Engine {
	r := gin.New()

	authH := NewAuthHandler(testDB)
	jobH := NewJobHandler(service.NewJobService(testDB, nil, nil, nil))
	pipelineH := NewPipelineHandler(service.NewPipelineService(testDB))

	r.POST("/api/admin/auth/login", authH.Login)

	admin := r.Group("/api/admin", middleware.JWTAuth())
	{
		admin.GET("/pipelines", pipelineH.List)
		admin.POST("/pipelines", pipelineH.Create)
		admin.GET("/pipelines/:id", pipelineH.Get)
		admin.DELETE("/pipelines/:id", pipelineH.Delete)

		admin.GET("/jobs", jobH.List)
		admin.POST("/jobs", jobH.Submit)
		admin.GET("/jobs/:id", jobH.GetStatus)
	}

	return r
}

func doRequest(method, path string, body any, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func getToken(t *testing.T) string {
	t.Helper()
	w := doRequest("POST", "/api/admin/auth/login", map[string]string{
		"username": "admin",
		"password": "testpass",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.Token
}

// --- Auth Tests ---

func TestLogin_Success(t *testing.T) {
	w := doRequest("POST", "/api/admin/auth/login", map[string]string{
		"username": "admin",
		"password": "testpass",
	}, "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["token"] == nil || data["token"] == "" {
		t.Error("expected token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	w := doRequest("POST", "/api/admin/auth/login", map[string]string{
		"username": "admin",
		"password": "wrongpass",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_MissingParams(t *testing.T) {
	w := doRequest("POST", "/api/admin/auth/login", map[string]string{
		"username": "admin",
	}, "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	w := doRequest("POST", "/api/admin/auth/login", map[string]string{
		"username": "nobody",
		"password": "testpass",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- JWT Middleware Tests ---

func TestJWT_NoToken(t *testing.T) {
	w := doRequest("GET", "/api/admin/pipelines", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	w := doRequest("GET", "/api/admin/pipelines", nil, "invalid.token.here")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWT_ValidToken(t *testing.T) {
	token := getToken(t)
	w := doRequest("GET", "/api/admin/pipelines", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- Pipeline Tests ---

func TestPipeline_CRUD(t *testing.T) {
	token := getToken(t)

	// Create
	w := doRequest("POST", "/api/admin/pipelines", map[string]any{
		"name":        "test-pipeline-" + t.Name(),
		"description": "test desc",
		"active":      true,
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("create: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var createResp struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	id := createResp.Data.ID
	if id == 0 {
		t.Fatal("expected pipeline ID > 0")
	}

	// Get
	w = doRequest("GET", "/api/admin/pipelines/"+itoa(id), nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("get: expected 200, got %d", w.Code)
	}

	// List
	w = doRequest("GET", "/api/admin/pipelines", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("list: expected 200, got %d", w.Code)
	}

	// Delete
	w = doRequest("DELETE", "/api/admin/pipelines/"+itoa(id), nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("delete: expected 200, got %d", w.Code)
	}

	// Get after delete → 404
	w = doRequest("GET", "/api/admin/pipelines/"+itoa(id), nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("get deleted: expected 404, got %d", w.Code)
	}
}

// --- Job Tests ---

func TestJob_Submit_MissingType(t *testing.T) {
	token := getToken(t)
	w := doRequest("POST", "/api/admin/jobs", map[string]any{
		"content":     "hello",
		"pipeline_id": 1,
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestJob_Submit_ImageMissingURL(t *testing.T) {
	token := getToken(t)
	w := doRequest("POST", "/api/admin/jobs", map[string]any{
		"type":        "image",
		"pipeline_id": 1,
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestJob_Submit_TextMissingContent(t *testing.T) {
	token := getToken(t)
	w := doRequest("POST", "/api/admin/jobs", map[string]any{
		"type":        "text",
		"pipeline_id": 1,
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestJob_GetStatus_NotFound(t *testing.T) {
	token := getToken(t)
	w := doRequest("GET", "/api/admin/jobs/nonexistent-uuid", nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestJob_List(t *testing.T) {
	token := getToken(t)
	w := doRequest("GET", "/api/admin/jobs?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Data struct {
			List  []any `json:"list"`
			Total int   `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.List == nil {
		t.Error("expected list to be non-nil")
	}
}

func itoa(n uint) string {
	return fmt.Sprintf("%d", n)
}
