package util_test

import (
	"encoding/base64"
	"imapsync-user/internal/util"
	"testing"
)

func TestGenerateRandomAESKey(t *testing.T) {
	encoded, raw, err := util.GenerateRandomAESKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw) != 32 {
		t.Fatalf("expected 32-byte key, got %d bytes", len(raw))
	}
	if encoded == "" {
		t.Fatal("expected non-empty base64 encoded key")
	}
}

func TestEncryptDecryptAESGCM_RoundTrip(t *testing.T) {
	_, key, err := util.GenerateRandomAESKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plainText := "super-secret-password"
	encrypted, err := util.EncryptAESGCM(plainText, key)
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}
	if encrypted == "" {
		t.Fatal("expected non-empty ciphertext")
	}

	decrypted, err := util.DecryptAESGCM(encrypted, key)
	if err != nil {
		t.Fatalf("unexpected decrypt error: %v", err)
	}
	if decrypted != plainText {
		t.Fatalf("expected %q, got %q", plainText, decrypted)
	}
}

func TestEncryptAESGCM_DifferentCipherTextEachTime(t *testing.T) {
	_, key, _ := util.GenerateRandomAESKey()

	first, err := util.EncryptAESGCM("same-input", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := util.EncryptAESGCM("same-input", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first == second {
		t.Fatal("expected different ciphertexts due to random nonce")
	}
}

func TestEncryptAESGCM_InvalidKeySize(t *testing.T) {
	_, err := util.EncryptAESGCM("data", []byte("too-short"))
	if err == nil {
		t.Fatal("expected error for invalid AES key size")
	}
}

func TestDecryptAESGCM_InvalidBase64(t *testing.T) {
	_, key, _ := util.GenerateRandomAESKey()
	_, err := util.DecryptAESGCM("not-valid-base64!!", key)
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}

func TestDecryptAESGCM_ShorterThanNonceSize(t *testing.T) {
	_, key, _ := util.GenerateRandomAESKey()
	tooShort := base64.StdEncoding.EncodeToString([]byte("ab"))

	_, err := util.DecryptAESGCM(tooShort, key)
	if err == nil {
		t.Fatal("expected error for ciphertext shorter than the GCM nonce size")
	}
}

func TestDecryptAESGCM_WrongKeyFailsAuth(t *testing.T) {
	_, key, _ := util.GenerateRandomAESKey()
	_, otherKey, _ := util.GenerateRandomAESKey()

	encrypted, err := util.EncryptAESGCM("secret", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = util.DecryptAESGCM(encrypted, otherKey)
	if err == nil {
		t.Fatal("expected authentication error when decrypting with the wrong key")
	}
}

func TestBase64Decode(t *testing.T) {
	decoded := util.Base64Decode("aGVsbG8=")
	if string(decoded) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", string(decoded))
	}
}

func TestBase64Decode_Invalid(t *testing.T) {
	decoded := util.Base64Decode("not base64 at all!!")
	if len(decoded) != 0 {
		t.Fatalf("expected no decoded bytes for invalid base64 input, got %v", decoded)
	}
}

func TestEncryptDecryptAESGCMStr_RoundTrip(t *testing.T) {
	_, key, _ := util.GenerateRandomAESKey()

	enc := util.EncryptAESGCMStr("hello world", key)
	if enc == "" {
		t.Fatal("expected non-empty ciphertext")
	}

	dec := util.DecryptAESGCMStr(enc, key)
	if dec != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", dec)
	}
}
