package service

import (
	"testing"
	"time"

	"github.com/mirainya/nexus/internal/model"
)

func seedAPIKeyForTest(t *testing.T) model.APIKey {
	t.Helper()
	ak := model.APIKey{Name: "test-ak-" + t.Name(), Key: "nxk_test_" + t.Name(), Active: true}
	if err := testDB.Create(&ak).Error; err != nil {
		t.Fatalf("seed api key: %v", err)
	}
	return ak
}

func TestAPIKeyService_Create(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	resp, err := svc.Create(APIKeyCreateRequest{Name: "my-key"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.ID == 0 {
		t.Error("expected ID > 0")
	}
	if resp.Key == "" {
		t.Error("expected key to be generated")
	}
	if resp.Key[:4] != "nxk_" {
		t.Errorf("expected key prefix 'nxk_', got %q", resp.Key[:4])
	}
	if !resp.Active {
		t.Error("expected active=true")
	}
}

func TestAPIKeyService_Create_WithExpiry(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	resp, err := svc.Create(APIKeyCreateRequest{
		Name:      "expiring-key",
		ExpiresAt: "2030-12-31",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.ExpiresAt == nil {
		t.Fatal("expected expires_at to be set")
	}
	if resp.ExpiresAt.Year() != 2030 {
		t.Errorf("expected year 2030, got %d", resp.ExpiresAt.Year())
	}
}

func TestAPIKeyService_Create_InvalidExpiry(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	_, err := svc.Create(APIKeyCreateRequest{
		Name:      "bad-expiry",
		ExpiresAt: "not-a-date",
	})
	if err == nil {
		t.Error("expected error for invalid expires_at")
	}
}

func TestAPIKeyService_List(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	svc.Create(APIKeyCreateRequest{Name: "list-key-1"})
	svc.Create(APIKeyCreateRequest{Name: "list-key-2"})

	list, err := svc.List(0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) < 2 {
		t.Errorf("expected >= 2 keys, got %d", len(list))
	}
	for _, k := range list {
		if k.Key[:4] == "nxk_" && len(k.Key) > 12 && k.Key[4] != '*' {
			t.Errorf("expected masked key in list, got %q", k.Key)
		}
	}
}

func TestAPIKeyService_Update(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	created, _ := svc.Create(APIKeyCreateRequest{Name: "update-key"})

	newLimit := 100
	resp, err := svc.Update(created.ID, APIKeyUpdateRequest{
		Name:       "renamed-key",
		DailyLimit: &newLimit,
	}, 0)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if resp.Name != "renamed-key" {
		t.Errorf("expected 'renamed-key', got %q", resp.Name)
	}
	if resp.DailyLimit != 100 {
		t.Errorf("expected daily_limit 100, got %d", resp.DailyLimit)
	}
}

func TestAPIKeyService_Update_ClearExpiry(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	created, _ := svc.Create(APIKeyCreateRequest{Name: "clear-expiry", ExpiresAt: "2030-01-01"})

	resp, err := svc.Update(created.ID, APIKeyUpdateRequest{ExpiresAt: "null"}, 0)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if resp.ExpiresAt != nil {
		t.Error("expected expires_at to be nil")
	}
}

func TestAPIKeyService_Delete(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	created, _ := svc.Create(APIKeyCreateRequest{Name: "delete-key"})

	if err := svc.Delete(created.ID, 0); err != nil {
		t.Fatalf("delete: %v", err)
	}

	list, _ := svc.List(0)
	for _, k := range list {
		if k.ID == created.ID {
			t.Error("expected key to be deleted")
		}
	}
}

func TestAPIKeyService_GetUsage(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	created, _ := svc.Create(APIKeyCreateRequest{Name: "usage-key"})

	today := time.Now().Format("2006-01-02")
	testDB.Create(&model.APIUsage{APIKeyID: created.ID, Date: today, Requests: 10, Tokens: 500})

	usages, err := svc.GetUsage(created.ID, 7, 0)
	if err != nil {
		t.Fatalf("get usage: %v", err)
	}
	if len(usages) != 1 {
		t.Fatalf("expected 1 usage record, got %d", len(usages))
	}
	if usages[0].Requests != 10 {
		t.Errorf("expected 10 requests, got %d", usages[0].Requests)
	}
	if usages[0].Tokens != 500 {
		t.Errorf("expected 500 tokens, got %d", usages[0].Tokens)
	}
}

func TestAPIKeyService_GetUsage_Empty(t *testing.T) {
	svc := NewAPIKeyService(testDB)
	created, _ := svc.Create(APIKeyCreateRequest{Name: "no-usage-key"})

	usages, err := svc.GetUsage(created.ID, 30, 0)
	if err != nil {
		t.Fatalf("get usage: %v", err)
	}
	if len(usages) != 0 {
		t.Errorf("expected 0 usage records, got %d", len(usages))
	}
}
