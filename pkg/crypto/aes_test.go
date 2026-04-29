package crypto

import (
	"os"
	"testing"

	"github.com/mirainya/nexus/pkg/config"
)

func TestMain(m *testing.M) {
	config.C = &config.Config{
		Server: config.ServerConfig{
			CredentialSecret: "test-credential-secret-key-32ch!",
		},
	}
	os.Exit(m.Run())
}

func TestEncryptDecrypt(t *testing.T) {
	plain := "sk-abc123456789xyz"
	encrypted, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == plain {
		t.Error("encrypted should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decrypted != plain {
		t.Errorf("expected %q, got %q", plain, decrypted)
	}
}

func TestEncryptDifferentNonce(t *testing.T) {
	plain := "sk-test-key"
	e1, _ := Encrypt(plain)
	e2, _ := Encrypt(plain)
	if e1 == e2 {
		t.Error("two encryptions of same plaintext should produce different ciphertext")
	}
	d1, _ := Decrypt(e1)
	d2, _ := Decrypt(e2)
	if d1 != plain || d2 != plain {
		t.Error("both should decrypt to original")
	}
}

func TestDecryptInvalid(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	_, err = Decrypt("YWJj")
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestDecryptTampered(t *testing.T) {
	encrypted, _ := Encrypt("secret")
	tampered := encrypted[:len(encrypted)-2] + "XX"
	_, err := Decrypt(tampered)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestShortSecret(t *testing.T) {
	orig := config.C.Server.CredentialSecret
	config.C.Server.CredentialSecret = "short"
	_, err := Encrypt("test")
	if err == nil {
		t.Error("expected error for short secret")
	}
	config.C.Server.CredentialSecret = orig
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"sk-abc123456789xyz", "sk-a****9xyz"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234****6789"},
	}
	for _, tt := range tests {
		got := MaskKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
