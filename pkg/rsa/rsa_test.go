/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rsa

import (
	"crypto/rand"
	_rsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// generateTestKeyPair generates a test RSA key pair
func generateTestKeyPair(bits int) (string, string, error) {
	// Generate private key
	privateKey, err := _rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), string(privateKeyPEM), nil
}

// TestSplit tests the split helper function
func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		buf      []byte
		lim      int
		expected int // expected number of chunks
	}{
		{
			name:     "Exact division",
			buf:      []byte("123456789012"),
			lim:      4,
			expected: 3,
		},
		{
			name:     "With remainder",
			buf:      []byte("1234567890"),
			lim:      3,
			expected: 4,
		},
		{
			name:     "Single chunk",
			buf:      []byte("12345"),
			lim:      10,
			expected: 1,
		},
		{
			name:     "Empty buffer",
			buf:      []byte{},
			lim:      5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := split(tt.buf, tt.lim)
			if len(chunks) != tt.expected {
				t.Errorf("expected %d chunks, got %d", tt.expected, len(chunks))
			}

			// Verify all data is preserved
			var reconstructed []byte
			for _, chunk := range chunks {
				reconstructed = append(reconstructed, chunk...)
			}
			if string(reconstructed) != string(tt.buf) {
				t.Error("split data doesn't match original")
			}
		})
	}
}

// TestRsaEncryptDecrypt tests basic encryption and decryption
func TestRsaEncryptDecrypt(t *testing.T) {
	// Generate test key pair (1024 bits for faster testing)
	pubKey, privKey, err := generateTestKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "Short text",
			plaintext: "Hello, World!",
		},
		{
			name:      "Empty string",
			plaintext: "",
		},
		{
			name:      "Special characters",
			plaintext: "Test Chinese characters!@#$%^&*()",
		},
		{
			name:      "Long text",
			plaintext: "This is a longer text that should be split into multiple chunks during encryption process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := RsaEncrypt([]byte(tt.plaintext), pubKey)
			if err != nil {
				t.Fatalf("RsaEncrypt failed: %v", err)
			}

			if encrypted == "" && tt.plaintext != "" {
				t.Error("encrypted string is empty")
			}

			// Decrypt
			decrypted, err := RsaDecrypt(encrypted, privKey, 1024)
			if err != nil {
				t.Fatalf("RsaDecrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("decrypted text doesn't match original\nwant: %q\ngot:  %q", tt.plaintext, decrypted)
			}
		})
	}
}

// TestRsaEncrypt_InvalidPublicKey tests encryption with invalid public key
func TestRsaEncrypt_InvalidPublicKey(t *testing.T) {
	tests := []struct {
		name   string
		pubKey string
	}{
		{
			name:   "Empty key",
			pubKey: "",
		},
		{
			name:   "Invalid PEM format",
			pubKey: "This is not a valid PEM key",
		},
		{
			name:   "Invalid PEM block",
			pubKey: "-----BEGIN PUBLIC KEY-----\nInvalidData\n-----END PUBLIC KEY-----",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RsaEncrypt([]byte("test"), tt.pubKey)
			if err == nil {
				t.Error("expected error with invalid public key, got nil")
			}
		})
	}
}

// TestRsaDecrypt_InvalidPrivateKey tests decryption with invalid private key
func TestRsaDecrypt_InvalidPrivateKey(t *testing.T) {
	pubKey, _, err := generateTestKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	encrypted, err := RsaEncrypt([]byte("test"), pubKey)
	if err != nil {
		t.Fatalf("RsaEncrypt failed: %v", err)
	}

	tests := []struct {
		name    string
		privKey string
	}{
		{
			name:    "Empty key",
			privKey: "",
		},
		{
			name:    "Invalid PEM format",
			privKey: "This is not a valid PEM key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RsaDecrypt(encrypted, tt.privKey, 1024)
			if err == nil {
				t.Error("expected error with invalid private key, got nil")
			}
		})
	}
}

// TestRsaDecrypt_WrongKey tests decryption with wrong key
func TestRsaDecrypt_WrongKey(t *testing.T) {
	// Generate first key pair
	pubKey1, _, err := generateTestKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	// Encrypt with first public key
	encrypted, err := RsaEncrypt([]byte("secret message"), pubKey1)
	if err != nil {
		t.Fatalf("RsaEncrypt failed: %v", err)
	}

	// Generate second key pair
	_, privKey2, err := generateTestKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	// Try to decrypt with second private key (should fail)
	_, err = RsaDecrypt(encrypted, privKey2, 1024)
	if err == nil {
		t.Error("expected error when decrypting with wrong key, got nil")
	}
}

// TestRsaEncryptDecrypt_DifferentKeySizes tests with different RSA key sizes
func TestRsaEncryptDecrypt_DifferentKeySizes(t *testing.T) {
	keySizes := []int{1024, 2048}

	for _, keySize := range keySizes {
		t.Run(string(rune(keySize)), func(t *testing.T) {
			pubKey, privKey, err := generateTestKeyPair(keySize)
			if err != nil {
				t.Fatalf("Failed to generate %d-bit key pair: %v", keySize, err)
			}

			plaintext := "Test message for different key sizes"

			encrypted, err := RsaEncrypt([]byte(plaintext), pubKey)
			if err != nil {
				t.Fatalf("RsaEncrypt failed with %d-bit key: %v", keySize, err)
			}

			decrypted, err := RsaDecrypt(encrypted, privKey, keySize)
			if err != nil {
				t.Fatalf("RsaDecrypt failed with %d-bit key: %v", keySize, err)
			}

			if decrypted != plaintext {
				t.Errorf("decrypted text doesn't match for %d-bit key", keySize)
			}
		})
	}
}

// TestRsaEncryptDecrypt_LargeData tests with large data requiring chunking
func TestRsaEncryptDecrypt_LargeData(t *testing.T) {
	pubKey, privKey, err := generateTestKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create large text (should require multiple chunks)
	var largeText string
	for i := 0; i < 1000; i++ {
		largeText += "This is line " + string(rune(i)) + " of the large text. "
	}

	encrypted, err := RsaEncrypt([]byte(largeText), pubKey)
	if err != nil {
		t.Fatalf("RsaEncrypt failed: %v", err)
	}

	decrypted, err := RsaDecrypt(encrypted, privKey, 2048)
	if err != nil {
		t.Fatalf("RsaDecrypt failed: %v", err)
	}

	if decrypted != largeText {
		t.Error("decrypted large text doesn't match original")
	}
}

// TestRsaDecrypt_InvalidEncryptedData tests decryption with invalid data
func TestRsaDecrypt_InvalidEncryptedData(t *testing.T) {
	_, privKey, err := generateTestKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tests := []struct {
		name      string
		encrypted string
	}{
		{
			name:      "Invalid base64",
			encrypted: "!!!invalid base64!!!",
		},
		{
			name:      "Valid base64 but invalid encrypted data",
			encrypted: "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RsaDecrypt(tt.encrypted, privKey, 1024)
			if err == nil {
				t.Error("expected error with invalid encrypted data, got nil")
			}
		})
	}
}

// TestRsaEncrypt_EmptyData tests encrypting empty data
func TestRsaEncrypt_EmptyData(t *testing.T) {
	pubKey, privKey, err := generateTestKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	encrypted, err := RsaEncrypt([]byte(""), pubKey)
	if err != nil {
		t.Fatalf("RsaEncrypt failed: %v", err)
	}

	decrypted, err := RsaDecrypt(encrypted, privKey, 1024)
	if err != nil {
		t.Fatalf("RsaDecrypt failed: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}

// BenchmarkRsaEncrypt benchmarks RSA encryption
func BenchmarkRsaEncrypt(b *testing.B) {
	pubKey, _, err := generateTestKeyPair(2048)
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	plaintext := []byte("Benchmark test message for RSA encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RsaEncrypt(plaintext, pubKey)
	}
}

// BenchmarkRsaDecrypt benchmarks RSA decryption
func BenchmarkRsaDecrypt(b *testing.B) {
	pubKey, privKey, err := generateTestKeyPair(2048)
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	plaintext := []byte("Benchmark test message for RSA decryption")
	encrypted, err := RsaEncrypt(plaintext, pubKey)
	if err != nil {
		b.Fatalf("RsaEncrypt failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RsaDecrypt(encrypted, privKey, 2048)
	}
}

// BenchmarkRsaEncryptDecrypt benchmarks full encrypt/decrypt cycle
func BenchmarkRsaEncryptDecrypt(b *testing.B) {
	pubKey, privKey, err := generateTestKeyPair(2048)
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	plaintext := []byte("Benchmark test message for full cycle")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, _ := RsaEncrypt(plaintext, pubKey)
		_, _ = RsaDecrypt(encrypted, privKey, 2048)
	}
}
