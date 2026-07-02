package security

import (
	"errors"
	"testing"
)

// Issue #17: CheckHost enforces the network allowlist for any host:port (used by
// the grpc/ssh runners, not just HTTP).
func TestCheckHost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		allow   []string
		host    string
		wantErr bool
	}{
		{"empty allowlist permits all", nil, "evil.example.com:443", false},
		{"host match", []string{"api.example.com"}, "api.example.com:50051", false},
		{"host:port match", []string{"api.example.com:50051"}, "api.example.com:50051", false},
		{"bare host match", []string{"api.example.com"}, "api.example.com", false},
		{"denied host", []string{"api.example.com"}, "evil.example.com:22", true},
		{"denied bare host", []string{"api.example.com"}, "evil.example.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CheckHost(tt.allow, tt.host)
			if tt.wantErr != (err != nil) {
				t.Fatalf("CheckHost(%v, %q) err = %v, wantErr %v", tt.allow, tt.host, err, tt.wantErr)
			}
			if tt.wantErr {
				var pe *PolicyError
				if !errors.As(err, &pe) {
					t.Errorf("error type = %T, want *PolicyError", err)
				}
			}
		})
	}
}
