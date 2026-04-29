package service

import (
	"testing"
	"time"

	"github.com/mirainya/nexus/internal/model"
)

func TestUsageTracker_FirstRecord(t *testing.T) {
	tracker := NewUsageTracker(testDB)

	ak := seedAPIKeyForTest(t)
	tracker.Track(&ak.ID, 100)

	today := time.Now().Format("2006-01-02")
	var usage model.APIUsage
	if err := testDB.Where("api_key_id = ? AND date = ?", ak.ID, today).First(&usage).Error; err != nil {
		t.Fatalf("expected usage record: %v", err)
	}
	if usage.Requests != 1 {
		t.Errorf("expected 1 request, got %d", usage.Requests)
	}
	if usage.Tokens != 100 {
		t.Errorf("expected 100 tokens, got %d", usage.Tokens)
	}
}

func TestUsageTracker_Accumulate(t *testing.T) {
	tracker := NewUsageTracker(testDB)

	ak := seedAPIKeyForTest(t)
	tracker.Track(&ak.ID, 50)
	tracker.Track(&ak.ID, 75)

	today := time.Now().Format("2006-01-02")
	var usage model.APIUsage
	testDB.Where("api_key_id = ? AND date = ?", ak.ID, today).First(&usage)

	if usage.Requests != 2 {
		t.Errorf("expected 2 requests, got %d", usage.Requests)
	}
	if usage.Tokens != 125 {
		t.Errorf("expected 125 tokens, got %d", usage.Tokens)
	}
}

func TestUsageTracker_NilAPIKeyID(t *testing.T) {
	tracker := NewUsageTracker(testDB)
	tracker.Track(nil, 100)
}

func TestUsageTracker_ZeroTokens(t *testing.T) {
	tracker := NewUsageTracker(testDB)
	ak := seedAPIKeyForTest(t)
	tracker.Track(&ak.ID, 0)

	today := time.Now().Format("2006-01-02")
	var usage model.APIUsage
	err := testDB.Where("api_key_id = ? AND date = ?", ak.ID, today).First(&usage).Error
	if err == nil {
		t.Error("expected no record for zero tokens")
	}
}
