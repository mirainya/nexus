package service

import (
	"testing"
)

func TestCredentialService_Create(t *testing.T) {
	svc := NewCredentialService(testDB)

	ak := seedAPIKeyForTest(t)
	resp, err := svc.Create(CredentialCreateRequest{
		APIKeyID:     ak.ID,
		Name:         "test-cred",
		ProviderType: "openai",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-test-secret-key-12345",
		DefaultModel: "gpt-4o",
	}, 0)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.ID == 0 {
		t.Error("expected ID > 0")
	}
	if resp.MaskedKey == "sk-test-secret-key-12345" {
		t.Error("expected masked key, got plaintext")
	}
}

func TestCredentialService_GetDecrypted(t *testing.T) {
	svc := NewCredentialService(testDB)

	ak := seedAPIKeyForTest(t)
	created, _ := svc.Create(CredentialCreateRequest{
		APIKeyID:     ak.ID,
		Name:         "decrypt-test",
		ProviderType: "openai",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-decrypt-me-12345",
	}, 0)

	cred, plain, err := svc.GetDecrypted(created.ID)
	if err != nil {
		t.Fatalf("get decrypted: %v", err)
	}
	if plain != "sk-decrypt-me-12345" {
		t.Errorf("expected 'sk-decrypt-me-12345', got %q", plain)
	}
	if cred.ProviderType != "openai" {
		t.Errorf("expected provider 'openai', got %q", cred.ProviderType)
	}
}

func TestCredentialService_List(t *testing.T) {
	svc := NewCredentialService(testDB)

	list, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, c := range list {
		if len(c.MaskedKey) > 0 && c.MaskedKey[0] == 's' && c.MaskedKey[1] == 'k' && c.MaskedKey[2] == '-' {
			if len(c.MaskedKey) > 12 && c.MaskedKey[4] != '*' {
				t.Errorf("expected masked key, got %q", c.MaskedKey)
			}
		}
	}
}

func TestCredentialService_Update(t *testing.T) {
	svc := NewCredentialService(testDB)

	ak := seedAPIKeyForTest(t)
	created, _ := svc.Create(CredentialCreateRequest{
		APIKeyID:     ak.ID,
		Name:         "update-test",
		ProviderType: "openai",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-update-original",
	}, 0)

	updated, err := svc.Update(created.ID, CredentialUpdateRequest{
		Name:   "updated-name",
		APIKey: "sk-update-new-key-12345",
	}, 0)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "updated-name" {
		t.Errorf("expected 'updated-name', got %q", updated.Name)
	}

	_, plain, _ := svc.GetDecrypted(created.ID)
	if plain != "sk-update-new-key-12345" {
		t.Errorf("expected new key, got %q", plain)
	}
}

func TestCredentialService_Delete(t *testing.T) {
	svc := NewCredentialService(testDB)

	ak := seedAPIKeyForTest(t)
	created, _ := svc.Create(CredentialCreateRequest{
		APIKeyID:     ak.ID,
		Name:         "delete-test",
		ProviderType: "openai",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-delete-me",
	}, 0)

	if err := svc.Delete(created.ID, 0); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, _, err := svc.GetDecrypted(created.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestCredentialService_GetDecrypted_Inactive(t *testing.T) {
	svc := NewCredentialService(testDB)

	ak := seedAPIKeyForTest(t)
	created, _ := svc.Create(CredentialCreateRequest{
		APIKeyID:     ak.ID,
		Name:         "inactive-test",
		ProviderType: "openai",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-inactive-key",
	}, 0)

	active := false
	svc.Update(created.ID, CredentialUpdateRequest{Active: &active}, 0)

	_, _, err := svc.GetDecrypted(created.ID)
	if err == nil {
		t.Error("expected error for inactive credential")
	}
}
