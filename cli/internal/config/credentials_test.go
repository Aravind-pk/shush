package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCredentials_Expired(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration // relative to now
		want      bool
	}{
		{"expired_long_ago", -1 * time.Hour, true},
		{"expired_recently", -1 * time.Second, true},
		{"expires_very_soon_skew", 10 * time.Second, true}, // within the 30s skew
		{"expires_soon", 40 * time.Second, false},          // outside the 30s skew
		{"expires_later", 1 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credentials{
				ExpiresAt: time.Now().Add(tt.expiresIn),
			}
			if got := c.Expired(); got != tt.want {
				t.Errorf("Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Isolate tests
	tmpDir := t.TempDir()
	t.Setenv("SHUSH_CONFIG_DIR", tmpDir)

	// Initially, not logged in
	_, err := Load()
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("expected ErrNotLoggedIn, got %v", err)
	}

	// Prepare data
	// Using UTC and truncating to avoid timezone/precision mismatches during JSON encode/decode
	now := time.Now().UTC().Truncate(time.Second)
	cred := &Credentials{
		ServerURL:    "http://localhost:8080",
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresAt:    now,
	}

	// Save
	if err := Save(cred); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Path and permissions check
	path, err := Path()
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}

	// Load
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ServerURL != cred.ServerURL {
		t.Errorf("ServerURL got %q, want %q", loaded.ServerURL, cred.ServerURL)
	}
	if loaded.AccessToken != cred.AccessToken {
		t.Errorf("AccessToken got %q, want %q", loaded.AccessToken, cred.AccessToken)
	}
	if loaded.RefreshToken != cred.RefreshToken {
		t.Errorf("RefreshToken got %q, want %q", loaded.RefreshToken, cred.RefreshToken)
	}
	if !loaded.ExpiresAt.Equal(cred.ExpiresAt) {
		t.Errorf("ExpiresAt got %v, want %v", loaded.ExpiresAt, cred.ExpiresAt)
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SHUSH_CONFIG_DIR", tmpDir)

	// Idempotency: Clear when not exists
	if err := Clear(); err != nil {
		t.Fatalf("Clear on non-existent file failed: %v", err)
	}

	// Save
	if err := Save(&Credentials{ServerURL: "test"}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Clear back to non-existent
	if err := Clear(); err != nil {
		t.Fatalf("Clear on existing file failed: %v", err)
	}

	// Verify it's gone
	_, err := Load()
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("expected ErrNotLoggedIn after Clear, got %v", err)
	}
}

func TestDirAndPath(t *testing.T) {
	// We can't guarantee UserHomeDir works in all test environments,
	// but if it does, we can verify the default behavior.
	t.Setenv("SHUSH_CONFIG_DIR", "")
	home, err := os.UserHomeDir()
	if err == nil {
		dir, _ := Dir()
		if dir != filepath.Join(home, ".shush") {
			t.Errorf("Dir() = %q, want %q", dir, filepath.Join(home, ".shush"))
		}
	}

	// Now test the override
	t.Setenv("SHUSH_CONFIG_DIR", "/tmp/shush_test")
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir failed: %v", err)
	}
	if dir != "/tmp/shush_test" {
		t.Errorf("Dir() = %q, want %q", dir, "/tmp/shush_test")
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}
	expectedPath := filepath.Join("/tmp/shush_test", "credentials.json")
	if path != expectedPath {
		t.Errorf("Path() = %q, want %q", path, expectedPath)
	}
}
