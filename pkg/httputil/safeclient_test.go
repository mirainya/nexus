package httputil

import (
	"net"
	"testing"
)

func TestValidateURL_PublicIP(t *testing.T) {
	if err := ValidateURL("https://8.8.8.8/image.jpg"); err != nil {
		t.Errorf("expected public IP to pass, got %v", err)
	}
}

func TestValidateURL_Loopback(t *testing.T) {
	if err := ValidateURL("http://127.0.0.1/secret"); err == nil {
		t.Error("expected loopback to be rejected")
	}
}

func TestValidateURL_PrivateClassA(t *testing.T) {
	if err := ValidateURL("http://10.0.0.1/internal"); err == nil {
		t.Error("expected 10.x.x.x to be rejected")
	}
}

func TestValidateURL_PrivateClassB(t *testing.T) {
	if err := ValidateURL("http://172.16.0.1/internal"); err == nil {
		t.Error("expected 172.16.x.x to be rejected")
	}
}

func TestValidateURL_PrivateClassC(t *testing.T) {
	if err := ValidateURL("http://192.168.1.1/internal"); err == nil {
		t.Error("expected 192.168.x.x to be rejected")
	}
}

func TestValidateURL_LinkLocal(t *testing.T) {
	if err := ValidateURL("http://169.254.169.254/latest/meta-data"); err == nil {
		t.Error("expected link-local (cloud metadata) to be rejected")
	}
}

func TestValidateURL_IPv6Loopback(t *testing.T) {
	if err := ValidateURL("http://[::1]/secret"); err == nil {
		t.Error("expected IPv6 loopback to be rejected")
	}
}

func TestValidateURL_BadScheme(t *testing.T) {
	if err := ValidateURL("ftp://example.com/file"); err == nil {
		t.Error("expected non-http(s) scheme to be rejected")
	}
	if err := ValidateURL("file:///etc/passwd"); err == nil {
		t.Error("expected file:// scheme to be rejected")
	}
}

func TestValidateURL_EmptyHost(t *testing.T) {
	if err := ValidateURL("http:///path"); err == nil {
		t.Error("expected empty host to be rejected")
	}
}

func TestValidateURL_ValidDomain(t *testing.T) {
	if err := ValidateURL("https://example.com/image.jpg"); err != nil {
		t.Errorf("expected public domain to pass, got %v", err)
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip      string
		private bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"::1", true},
	}
	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		got := isPrivateIP(ip)
		if got != tt.private {
			t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.private)
		}
	}
}
