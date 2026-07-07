package crypto

import "testing"

func TestSecretBoxEncryptDecrypt(t *testing.T) {
	box, err := NewSecretBox("unit-test-secret")
	if err != nil {
		t.Fatalf("new secret box: %v", err)
	}
	ciphertext, err := box.Encrypt("secret-value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if ciphertext == "" || ciphertext == "secret-value" {
		t.Fatal("ciphertext must be non-empty and not equal plaintext")
	}
	plaintext, err := box.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plaintext != "secret-value" {
		t.Fatalf("unexpected plaintext: %s", plaintext)
	}
}

func TestSecretBoxRequiresKey(t *testing.T) {
	if _, err := NewSecretBox(""); err == nil {
		t.Fatal("expected empty encryption key to fail")
	}
}
