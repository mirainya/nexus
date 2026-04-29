package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupWebhookTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&model.WebhookLog{})
	return db
}

func newTestWebhookService(t *testing.T, db *gorm.DB) *WebhookService {
	t.Helper()
	svc := NewWebhookService(db)
	svc.validateURL = func(string) error { return nil }
	return svc
}

func TestComputeHMAC(t *testing.T) {
	body := []byte(`{"event":"job.completed","job_id":1}`)
	secret := "test-secret-key"

	sig := computeHMAC(body, secret)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if sig != expected {
		t.Errorf("HMAC mismatch: got %s, want %s", sig, expected)
	}
}

func TestWebhookSend_Success(t *testing.T) {
	var receivedSig string
	var receivedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("X-Nexus-Signature")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	db := setupWebhookTestDB(t)
	svc := newTestWebhookService(t, db)

	payload := WebhookPayload{
		Event:   "job.completed",
		JobID:   1,
		JobUUID: "test-uuid",
		Status:  "completed",
	}
	svc.Send(srv.URL, payload, "my-secret")

	if receivedSig == "" {
		t.Fatal("expected X-Nexus-Signature header")
	}

	expectedSig := computeHMAC(receivedBody, "my-secret")
	if receivedSig != expectedSig {
		t.Errorf("signature mismatch: got %s, want %s", receivedSig, expectedSig)
	}

	var p WebhookPayload
	if err := json.Unmarshal(receivedBody, &p); err != nil {
		t.Fatal(err)
	}
	if p.Event != "job.completed" || p.JobID != 1 {
		t.Errorf("unexpected payload: %+v", p)
	}

	var log model.WebhookLog
	db.First(&log)
	if log.Status != "success" || log.Attempts != 1 {
		t.Errorf("unexpected log: status=%s attempts=%d", log.Status, log.Attempts)
	}
}

func TestWebhookSend_RetryThenSuccess(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	db := setupWebhookTestDB(t)
	svc := newTestWebhookService(t, db)

	payload := WebhookPayload{Event: "job.failed", JobID: 2, JobUUID: "retry-uuid", Status: "failed", Error: "some error"}
	svc.Send(srv.URL, payload, "secret")

	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}

	var log model.WebhookLog
	db.First(&log)
	if log.Status != "success" {
		t.Errorf("expected success after retry, got %s", log.Status)
	}
}

func TestWebhookSend_AllRetriesFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	db := setupWebhookTestDB(t)
	svc := newTestWebhookService(t, db)

	payload := WebhookPayload{Event: "job.failed", JobID: 3, JobUUID: "fail-uuid", Status: "failed"}
	svc.Send(srv.URL, payload, "secret")

	var log model.WebhookLog
	db.First(&log)
	if log.Status != "failed" {
		t.Errorf("expected failed, got %s", log.Status)
	}
	if log.Attempts != 4 {
		t.Errorf("expected 4 attempts, got %d", log.Attempts)
	}
}

func TestWebhookSend_SSRFBlocked(t *testing.T) {
	db := setupWebhookTestDB(t)
	svc := NewWebhookService(db)

	payload := WebhookPayload{Event: "job.completed", JobID: 4, JobUUID: "ssrf-uuid", Status: "completed"}
	svc.Send("http://127.0.0.1:8080/callback", payload, "secret")

	var log model.WebhookLog
	db.First(&log)
	if log.Status != "failed" {
		t.Errorf("expected failed for private IP, got %s", log.Status)
	}
}

func TestWebhookSend_EmptyURL(t *testing.T) {
	db := setupWebhookTestDB(t)
	svc := NewWebhookService(db)

	payload := WebhookPayload{Event: "job.completed", JobID: 5, JobUUID: "empty-uuid", Status: "completed"}
	svc.Send("", payload, "secret")

	var count int64
	db.Model(&model.WebhookLog{}).Where("status = 'success'").Count(&count)
	if count != 0 {
		t.Error("expected no success log for empty URL")
	}
}

func TestWebhookList(t *testing.T) {
	db := setupWebhookTestDB(t)
	svc := NewWebhookService(db)

	for i := 0; i < 5; i++ {
		db.Create(&model.WebhookLog{JobID: uint(i + 1), Event: "job.completed", Status: "success"})
	}

	logs, total, err := svc.List(1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
}
