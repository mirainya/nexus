package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/httputil"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type httpPostFunc func(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error)

type WebhookService struct {
	db          *gorm.DB
	validateURL func(string) error
	postFunc    httpPostFunc
}

type WebhookPayload struct {
	Event     string `json:"event"`
	JobID     uint   `json:"job_id"`
	JobUUID   string `json:"job_uuid"`
	Status    string `json:"status"`
	Result    any    `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db, validateURL: httputil.ValidateURL, postFunc: httputil.SafePost}
}

func (s *WebhookService) Send(callbackURL string, payload WebhookPayload, secret string) {
	if err := s.validateURL(callbackURL); err != nil {
		s.saveLog(payload.JobID, callbackURL, payload.Event, "failed", 0, err.Error(), 0)
		logger.Warn("webhook URL validation failed", zap.String("url", callbackURL), zap.Error(err))
		return
	}

	const maxRetries = 3
	var lastErr error
	var lastStatusCode int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}
		statusCode, err := s.doSend(callbackURL, payload, secret)
		lastStatusCode = statusCode
		if err == nil {
			s.saveLog(payload.JobID, callbackURL, payload.Event, "success", statusCode, "", attempt+1)
			return
		}
		lastErr = err
		logger.Warn("webhook delivery failed",
			zap.String("url", callbackURL),
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}

	s.saveLog(payload.JobID, callbackURL, payload.Event, "failed", lastStatusCode, lastErr.Error(), maxRetries+1)
}

func (s *WebhookService) doSend(callbackURL string, payload WebhookPayload, secret string) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	sig := computeHMAC(body, secret)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	headers := map[string]string{
		"Content-Type":      "application/json",
		"X-Nexus-Signature": sig,
	}
	resp, err := s.postFunc(ctx, callbackURL, bytes.NewReader(body), headers)
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, nil
	}
	return resp.StatusCode, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

func computeHMAC(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *WebhookService) saveLog(jobID uint, url, event, status string, statusCode int, errMsg string, attempts int) {
	log := model.WebhookLog{
		JobID:      jobID,
		URL:        url,
		Event:      event,
		Status:     status,
		StatusCode: statusCode,
		Error:      errMsg,
		Attempts:   attempts,
	}
	if err := s.db.Create(&log).Error; err != nil {
		logger.Warn("failed to save webhook log", zap.Uint("job_id", jobID), zap.Error(err))
	}
}

func (s *WebhookService) List(page, pageSize int) ([]model.WebhookLog, int64, error) {
	var logs []model.WebhookLog
	var total int64
	q := s.db.Model(&model.WebhookLog{})
	q.Count(&total)
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
