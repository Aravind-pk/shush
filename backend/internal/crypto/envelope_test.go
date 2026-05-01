package crypto

import (
	"bytes"
	"testing"
)

var testKEK = bytes.Repeat([]byte("k"), 32) // 32 bytes of 'k' — fine for tests

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plaintext := "super-secret-database-password"

	es, err := Encrypt(testKEK, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	got, err := Decrypt(testKEK, es)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if got != plaintext {
		t.Errorf("got %q, want %q", got, plaintext)
	}
}

func TestEachEncryptionProducesUniqueCiphertext(t *testing.T) {
	// Because we use a random nonce each time, the same plaintext encrypted
	// twice should produce different ciphertexts. If they matched, an attacker
	// could detect that two secrets have the same value.
	es1, _ := Encrypt(testKEK, "same-value")
	es2, _ := Encrypt(testKEK, "same-value")

	if bytes.Equal(es1.EncryptedValue, es2.EncryptedValue) {
		t.Error("two encryptions of the same value produced the same ciphertext — nonce reuse bug")
	}
}

func TestDecryptFailsWithWrongKEK(t *testing.T) {
	es, _ := Encrypt(testKEK, "secret")

	wrongKEK := bytes.Repeat([]byte("x"), 32)
	_, err := Decrypt(wrongKEK, es)
	if err == nil {
		t.Error("expected error when decrypting with wrong KEK, got nil")
	}
}

func TestDecryptFailsWithTamperedCiphertext(t *testing.T) {
	es, _ := Encrypt(testKEK, "secret")

	// Flip a bit in the ciphertext — GCM's authentication tag should catch this.
	es.EncryptedValue[0] ^= 0xFF

	_, err := Decrypt(testKEK, es)
	if err == nil {
		t.Error("expected error when decrypting tampered ciphertext, got nil")
	}
}

func TestInvalidKEKLength(t *testing.T) {
	shortKEK := []byte("too-short")

	_, err := Encrypt(shortKEK, "secret")
	if err == nil {
		t.Error("expected error for short KEK, got nil")
	}
}
