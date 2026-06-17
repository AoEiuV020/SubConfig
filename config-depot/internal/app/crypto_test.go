package app

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestDecryptMatchesWorkflowOpenSSL(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	ciphertext := mustDecodeBase64(t, "ceFxS7ro9hgTgz6P5+ThPHLgbEZlztkzY+dO0BznT1Q=")

	plaintext, err := DecryptAES256CBC(ciphertext, key, DefaultUploadIV)
	if err != nil {
		t.Fatalf("decrypt openssl ciphertext: %v", err)
	}

	if string(plaintext) != "hello config-depot\n" {
		t.Fatalf("plaintext = %q", plaintext)
	}
}

func TestEncryptMatchesWorkflowOpenSSL(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	want := mustDecodeBase64(t, "ceFxS7ro9hgTgz6P5+ThPHLgbEZlztkzY+dO0BznT1Q=")

	got, err := EncryptAES256CBC([]byte("hello config-depot\n"), key, DefaultUploadIV)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ciphertext = %x, want %x", got, want)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	plaintext := []byte("config archive bytes")

	ciphertext, err := EncryptAES256CBC(plaintext, key, DefaultUploadIV)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	decrypted, err := DecryptAES256CBC(ciphertext, key, DefaultUploadIV)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted = %q", decrypted)
	}
}

func mustDecodeBase64(t *testing.T, value string) []byte {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatalf("decode base64 %q: %v", value, err)
	}
	return decoded
}
