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
	"time"

	"github.com/Aravind-pk/shush/backend/internal/crypto"
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

// Project is the domain view of a project row.
type Project struct {
	ID        string
	Name      string
	OwnerID   string
	CreatedAt string
}

// EnsureUser find-or-creates a local user row for a Clerk user id and returns
// our internal user UUID. Called on each authenticated request to map the
// Clerk identity onto a foreign-keyable users.id.
//
// Clerk session tokens don't carry an email by default, so we synthesize a
// unique placeholder from the Clerk id to satisfy the NOT NULL/UNIQUE email
// column. (A later pass can backfill the real email from the Clerk API.)
func (s *Store) EnsureUser(ctx context.Context, clerkUserID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, clerk_user_id, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (clerk_user_id) DO UPDATE SET clerk_user_id = EXCLUDED.clerk_user_id
		RETURNING id`, clerkUserID+"@clerk.local", clerkUserID).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("store: ensure user: %w", err)
	}
	return id, nil
}

// CreateProject inserts a project owned by ownerID.
func (s *Store) CreateProject(ctx context.Context, ownerID, name string) (Project, error) {
	var p Project
	var created time.Time
	err := s.pool.QueryRow(ctx,
		`INSERT INTO projects (name, owner_id) VALUES ($1, $2)
		 RETURNING id, name, owner_id, created_at`, name, ownerID,
	).Scan(&p.ID, &p.Name, &p.OwnerID, &created)
	if err != nil {
		return Project{}, fmt.Errorf("store: create project: %w", err)
	}
	p.CreatedAt = created.Format(time.RFC3339)
	return p, nil
}

// ListProjects returns every project owned by ownerID, newest first.
func (s *Store) ListProjects(ctx context.Context, ownerID string) ([]Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, owner_id, created_at FROM projects
		 WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("store: query projects: %w", err)
	}
	defer rows.Close()

	var out []Project
	for rows.Next() {
		var p Project
		var created time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.OwnerID, &created); err != nil {
			return nil, fmt.Errorf("store: scan project: %w", err)
		}
		p.CreatedAt = created.Format(time.RFC3339)
		out = append(out, p)
	}
	return out, rows.Err()
}
