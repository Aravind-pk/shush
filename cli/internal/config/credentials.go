// Package config manages the CLI's local on-disk state, primarily the
// credentials file written after `shush login`.
//
// Credentials live at ~/.shush/credentials.json with mode 0600 (owner
// read/write only) inside a directory with mode 0700. This matches the
// pattern used by aws-cli, gcloud, and flyctl: plaintext JSON, protected
// by filesystem permissions, and kept short-lived via JWT expiry.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials is what we persist after a successful login.
//
// ServerURL is stored alongside the tokens so the CLI knows which Shush
// instance the token was issued by — important once self-hosted setups
// exist and one developer might talk to multiple servers.
type Credentials struct {
	ServerURL    string    `json:"server_url"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Expired reports whether the access token is past its expiry.
// A 30-second skew buffer avoids the case where we *just* checked
// "not expired" and then the request lands a moment too late.
func (c *Credentials) Expired() bool {
	return time.Now().Add(30 * time.Second).After(c.ExpiresAt)
}

// ErrNotLoggedIn is returned by Load when no credentials file exists.
// Callers can use errors.Is to distinguish "user hasn't logged in yet"
// from "credentials file is corrupted".
var ErrNotLoggedIn = errors.New("not logged in: run `shush login` first")

// dirName and fileName are split out so tests (later) can override the
// base directory by setting $HOME or $SHUSH_CONFIG_DIR.
const (
	dirName  = ".shush"
	fileName = "credentials.json"
)

// Dir returns the directory that holds the CLI's config (~/.shush).
// Honors $SHUSH_CONFIG_DIR for tests and unusual setups.
func Dir() (string, error) {
	if override := os.Getenv("SHUSH_CONFIG_DIR"); override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, dirName), nil
}

// Path returns the full path to credentials.json.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Load reads and parses the credentials file.
// Returns ErrNotLoggedIn if the file does not exist.
func Load() (*Credentials, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotLoggedIn
	}
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse credentials (file may be corrupt): %w", err)
	}
	return &c, nil
}

// Save writes credentials atomically with mode 0600.
//
// The atomic part matters: if we crashed mid-write, a half-written JSON
// file would brick the CLI. So we write to a temp file in the same
// directory, then rename — rename is atomic on POSIX filesystems.
func Save(c *Credentials) error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	// 0700 = only the owner can list/enter the directory.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}

	// Write to a temp file in the same directory so the final rename
	// stays on the same filesystem (rename across mounts is not atomic).
	tmp, err := os.CreateTemp(dir, "credentials-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// On any error after this point, clean up the temp file.
	cleanup := func() { _ = os.Remove(tmpPath) }

	// Tighten perms before writing the secret. CreateTemp defaults to
	// 0600 already on Unix, but we set it explicitly to be safe.
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}

	path := filepath.Join(dir, fileName)
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename into place: %w", err)
	}
	return nil
}

// Clear removes the credentials file. Used by `shush logout` (later).
// Missing-file is not an error — logout is idempotent.
func Clear() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}
