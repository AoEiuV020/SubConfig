package app

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestDecryptMatchesWorkflowOpenSSL(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	iv := mustDecodeBase64(t, "EJwC9OfO/fkuTvPax7YHeQ==")
	ciphertext := mustDecodeBase64(t, "ceFxS7ro9hgTgz6P5+ThPHLgbEZlztkzY+dO0BznT1Q=")

	plaintext, err := DecryptAES256CBC(ciphertext, key, iv)
	if err != nil {
		t.Fatalf("decrypt openssl ciphertext: %v", err)
	}

	if string(plaintext) != "hello config-depot\n" {
		t.Fatalf("plaintext = %q", plaintext)
	}
}

func TestEncryptMatchesWorkflowOpenSSL(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	iv := mustDecodeBase64(t, "EJwC9OfO/fkuTvPax7YHeQ==")
	want := mustDecodeBase64(t, "ceFxS7ro9hgTgz6P5+ThPHLgbEZlztkzY+dO0BznT1Q=")

	got, err := EncryptAES256CBC([]byte("hello config-depot\n"), key, iv)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ciphertext = %x, want %x", got, want)
	}
}

func TestUploadBundleEncryptDecryptRoundTrip(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	plaintext := []byte("config archive bytes")

	encrypted, err := EncryptUploadBundle(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	decrypted, err := DecryptUploadBundle(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted = %q", decrypted)
	}
}

func TestEncryptUploadBundleUsesFreshIV(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	plaintext := []byte("same archive bytes")

	first, err := EncryptUploadBundle(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	second, err := EncryptUploadBundle(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt again: %v", err)
	}
	if bytes.Equal(first, second) {
		t.Fatalf("encrypted payloads should differ")
	}
}

func TestDecryptUploadBundleReadsPrefixedIV(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	iv := mustDecodeBase64(t, "EJwC9OfO/fkuTvPax7YHeQ==")
	plaintext := []byte("config archive bytes")

	ciphertext, err := EncryptAES256CBC(plaintext, key, iv)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	encrypted := append(append([]byte{}, iv...), ciphertext...)

	decrypted, err := DecryptUploadBundle(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted = %q", decrypted)
	}
}

func TestDecryptUploadBundleRejectsMissingIV(t *testing.T) {
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")

	for _, encrypted := range [][]byte{nil, bytes.Repeat([]byte{0x01}, 15)} {
		if _, err := DecryptUploadBundle(encrypted, key); err == nil {
			t.Fatalf("decrypt should reject payload length %d", len(encrypted))
		}
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
