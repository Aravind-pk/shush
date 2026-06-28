// Package store is the PostgreSQL persistence layer for Shush.
//
// It owns a pgx connection pool and exposes typed methods for the domain
// (secrets, and demo bootstrap helpers for now). Encryption happens here, at
// the boundary: plaintext goes in / comes out of these methods, but only
// ciphertext is ever written to or read from the database.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/Aravind-pk/shush/backend/internal/crypto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pgx pool plus the master KEK used to wrap/unwrap secret DEKs.
type Store struct {
	pool *pgxpool.Pool
	kek  []byte
}

// Secret is the decrypted, domain-level view of a stored secret.
type Secret struct {
	Key     string
	Value   string
	Version int32
}

// New opens a pooled connection to dsn and verifies it with a ping.
// kek must be exactly 32 bytes (AES-256) — same requirement as the crypto pkg.
func New(ctx context.Context, dsn string, kek []byte) (*Store, error) {
	if len(kek) != 32 {
		return nil, errors.New("store: MASTER_KEK must be exactly 32 bytes")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool, kek: kek}, nil
}

// Close releases the connection pool. Safe to defer in main.
func (s *Store) Close() { s.pool.Close() }

// ListSecrets returns every secret for a project+environment, decrypted.
//
// It selects only the four ciphertext/nonce columns (never a plaintext column —
// there isn't one), reconstructs the envelope, and decrypts each row.
func (s *Store) ListSecrets(ctx context.Context, projectID, env string) ([]Secret, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT key, encrypted_value, data_nonce, encrypted_dek, dek_nonce, version
		FROM secrets
		WHERE project_id = $1 AND environment = $2
		ORDER BY key`, projectID, env)
	if err != nil {
		return nil, fmt.Errorf("store: query secrets: %w", err)
	}
	defer rows.Close()

	var out []Secret
	for rows.Next() {
		var (
			key                 string
			encValue, dataNonce []byte
			encDEK, dekNonce    []byte
			version             int32
		)
		if err := rows.Scan(&key, &encValue, &dataNonce, &encDEK, &dekNonce, &version); err != nil {
			return nil, fmt.Errorf("store: scan secret: %w", err)
		}

		plaintext, err := crypto.Decrypt(s.kek, &crypto.EncryptedSecret{
			EncryptedValue: encValue,
			DataNonce:      dataNonce,
			EncryptedDEK:   encDEK,
			DEKNonce:       dekNonce,
		})
		if err != nil {
			return nil, fmt.Errorf("store: decrypt %q: %w", key, err)
		}
		out = append(out, Secret{Key: key, Value: plaintext, Version: version})
	}
	return out, rows.Err()
}

// PutSecret encrypts value and upserts it. A new key starts at version 1; an
// existing key has its ciphertext replaced and version bumped. Returns the row
// id and the resulting version.
func (s *Store) PutSecret(ctx context.Context, projectID, env, key, value, createdBy string) (id string, version int32, err error) {
	enc, err := crypto.Encrypt(s.kek, value)
	if err != nil {
		return "", 0, fmt.Errorf("store: encrypt: %w", err)
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO secrets
			(project_id, environment, key, encrypted_value, data_nonce, encrypted_dek, dek_nonce, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (project_id, environment, key) DO UPDATE SET
			encrypted_value = EXCLUDED.encrypted_value,
			data_nonce      = EXCLUDED.data_nonce,
			encrypted_dek   = EXCLUDED.encrypted_dek,
			dek_nonce       = EXCLUDED.dek_nonce,
			version         = secrets.version + 1,
			updated_at      = NOW()
		RETURNING id, version`,
		projectID, env, key, enc.EncryptedValue, enc.DataNonce, enc.EncryptedDEK, enc.DEKNonce, createdBy,
	).Scan(&id, &version)
	if err != nil {
		return "", 0, fmt.Errorf("store: upsert secret: %w", err)
	}
	return id, version, nil
}

// EnsureDemoData find-or-creates a demo user and project so the secret RPCs
// have valid foreign keys to reference before auth/projects RPCs exist.
// Returns the demo user id and project id. Temporary — removed once Register
// and CreateProject are implemented.
func (s *Store) EnsureDemoData(ctx context.Context) (userID, projectID string, err error) {
	// users.email is UNIQUE, so a real upsert works here.
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ('demo@shush.local', 'x', 'admin')
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id`).Scan(&userID)
	if err != nil {
		return "", "", fmt.Errorf("store: ensure demo user: %w", err)
	}

	// projects has no unique constraint on name, so find-then-insert.
	err = s.pool.QueryRow(ctx,
		`SELECT id FROM projects WHERE name = 'demo' AND owner_id = $1 LIMIT 1`, userID,
	).Scan(&projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = s.pool.QueryRow(ctx,
			`INSERT INTO projects (name, owner_id) VALUES ('demo', $1) RETURNING id`, userID,
		).Scan(&projectID)
	}
	if err != nil {
		return "", "", fmt.Errorf("store: ensure demo project: %w", err)
	}
	return userID, projectID, nil
}
