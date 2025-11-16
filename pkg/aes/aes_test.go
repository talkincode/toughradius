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

package aes

import (
	"bytes"
	"strings"
	"testing"
)

const (
	// AES-256 key (32 bytes)
	key32 = "12345678123456781234567812345678"
	// AES-192 key (24 bytes)
	key24 = "123456781234567812345678"
	// AES-128 key (16 bytes)
	key16 = "1234567812345678"
)

// TestEncryptDecrypt tests basic encryption and decryption with different key sizes.
func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		key   string
	}{
		{
			name:  "AES-128 with short string",
			input: "hello",
			key:   key16,
		},
		{
			name:  "AES-192 with medium string",
			input: "hello world",
			key:   key24,
		},
		{
			name:  "AES-256 with long string",
			input: "The quick brown fox jumps over the lazy dog",
			key:   key32,
		},
		{
			name:  "Empty string",
			input: "",
			key:   key32,
		},
		{
			name:  "Single character",
			input: "a",
			key:   key16,
		},
		{
			name:  "Exactly one block (16 bytes)",
			input: "1234567890123456",
			key:   key32,
		},
		{
			name:  "Unicode characters",
			input: "你好世界 🌍",
			key:   key32,
		},
		{
			name:  "Special characters",
			input: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			key:   key24,
		},
		{
			name:  "Very long string",
			input: strings.Repeat("test", 100),
			key:   key32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test byte-level encryption/decryption
			encrypted, err := Encrypt([]byte(tt.input), tt.key)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := Decrypt(encrypted, tt.key)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal([]byte(tt.input), decrypted) {
				t.Errorf("Decrypt(Encrypt(%q)) = %q, want %q", tt.input, string(decrypted), tt.input)
			}
		})
	}
}

// TestEncryptToB64DecryptFromB64 tests base64 encryption and decryption.
func TestEncryptToB64DecryptFromB64(t *testing.T) {
	tests := []struct {
		name  string
		input string
		key   string
	}{
		{
			name:  "Simple password",
			input: "mypassword123",
			key:   key32,
		},
		{
			name:  "Empty string",
			input: "",
			key:   key16,
		},
		{
			name:  "Long password with special chars",
			input: "P@ssw0rd!2023#SecurePassword$%^&*()",
			key:   key24,
		},
		{
			name:  "RADIUS shared secret",
			input: "testing123",
			key:   key32,
		},
		{
			name:  "JSON config value",
			input: `{"api_key":"sk-1234567890abcdef","secret":"xyz"}`,
			key:   key32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptToB64(tt.input, tt.key)
			if err != nil {
				t.Fatalf("EncryptToB64() error = %v", err)
			}

			// Verify base64 output is not empty for non-empty input
			if tt.input != "" && encrypted == "" {
				t.Error("EncryptToB64() returned empty string for non-empty input")
			}

			decrypted, err := DecryptFromB64(encrypted, tt.key)
			if err != nil {
				t.Fatalf("DecryptFromB64() error = %v", err)
			}

			if decrypted != tt.input {
				t.Errorf("DecryptFromB64(EncryptToB64(%q)) = %q, want %q", tt.input, decrypted, tt.input)
			}
		})
	}
}

// TestEncryptWithInvalidKey tests encryption with invalid key sizes.
func TestEncryptWithInvalidKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "Too short key (15 bytes)",
			key:     "123456789012345",
			wantErr: true,
		},
		{
			name:    "Invalid key size (17 bytes)",
			key:     "12345678901234567",
			wantErr: true,
		},
		{
			name:    "Invalid key size (20 bytes)",
			key:     "12345678901234567890",
			wantErr: true,
		},
		{
			name:    "Valid AES-128 key",
			key:     key16,
			wantErr: false,
		},
		{
			name:    "Valid AES-192 key",
			key:     key24,
			wantErr: false,
		},
		{
			name:    "Valid AES-256 key",
			key:     key32,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Encrypt([]byte("test"), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDecryptWithWrongKey tests decryption with incorrect key.
func TestDecryptWithWrongKey(t *testing.T) {
	original := "sensitive data"
	correctKey := key32
	wrongKey := "00000000000000000000000000000000"

	encrypted, err := Encrypt([]byte(original), correctKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Decrypt with wrong key should fail
	decrypted, err := Decrypt(encrypted, wrongKey)
	if err == nil {
		// Even if no error, the decrypted data should not match original
		if string(decrypted) == original {
			t.Error("Decrypt with wrong key should not produce original data")
		}
	}
}

// TestDecryptFromB64WithInvalidBase64 tests decryption with invalid base64 input.
func TestDecryptFromB64WithInvalidBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Invalid base64 characters",
			input:   "not-valid-base64!@#$",
			wantErr: true,
		},
		{
			name:    "Invalid base64 padding",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true, // Empty base64 decodes to empty bytes, which fails in PKCS7UnPadding
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptFromB64(tt.input, key32)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFromB64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPKCS7Padding tests PKCS7 padding function.
func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		blocksize int
		wantLen   int
	}{
		{
			name:      "Empty input",
			input:     []byte{},
			blocksize: 16,
			wantLen:   16, // Should add full block of padding
		},
		{
			name:      "1 byte input",
			input:     []byte{0x01},
			blocksize: 16,
			wantLen:   16, // Should pad to block size
		},
		{
			name:      "15 bytes input",
			input:     bytes.Repeat([]byte{0x01}, 15),
			blocksize: 16,
			wantLen:   16, // Should add 1 byte of padding
		},
		{
			name:      "Exactly one block",
			input:     bytes.Repeat([]byte{0x01}, 16),
			blocksize: 16,
			wantLen:   32, // Should add full block of padding
		},
		{
			name:      "17 bytes input",
			input:     bytes.Repeat([]byte{0x01}, 17),
			blocksize: 16,
			wantLen:   32, // Should pad to next block boundary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PKCS7Padding(tt.input, tt.blocksize)
			if len(result) != tt.wantLen {
				t.Errorf("PKCS7Padding() length = %d, want %d", len(result), tt.wantLen)
			}

			// Verify padding value is correct
			if len(result) > 0 {
				paddingValue := result[len(result)-1]
				expectedPadding := byte(tt.wantLen - len(tt.input))
				if paddingValue != expectedPadding {
					t.Errorf("PKCS7Padding() padding value = %d, want %d", paddingValue, expectedPadding)
				}
			}
		})
	}
}

// TestPKCS7UnPadding tests PKCS7 unpadding function.
func TestPKCS7UnPadding(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "Empty input",
			input:   []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Valid padding - 1 byte of data",
			input:   []byte{0x01, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0x0F},
			want:    []byte{0x01},
			wantErr: false,
		},
		{
			name:    "Valid padding - full block",
			input:   []byte{0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10},
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "Invalid padding - exceeds length",
			input:   []byte{0x01, 0x02, 0x03, 0xFF},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PKCS7UnPadding(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PKCS7UnPadding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(result, tt.want) {
				t.Errorf("PKCS7UnPadding() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestEncryptDecryptConsistency verifies that multiple encryptions produce different ciphertexts
// but decrypt to the same plaintext (due to deterministic IV from key).
func TestEncryptDecryptConsistency(t *testing.T) {
	plaintext := "consistency test"
	key := key32

	// Encrypt the same plaintext twice
	encrypted1, err1 := Encrypt([]byte(plaintext), key)
	if err1 != nil {
		t.Fatalf("First Encrypt() error = %v", err1)
	}

	encrypted2, err2 := Encrypt([]byte(plaintext), key)
	if err2 != nil {
		t.Fatalf("Second Encrypt() error = %v", err2)
	}

	// With deterministic IV (key-based), encryptions should be identical
	if !bytes.Equal(encrypted1, encrypted2) {
		t.Error("Expected identical ciphertexts with same key and plaintext")
	}

	// Both should decrypt to original
	decrypted1, _ := Decrypt(encrypted1, key)
	decrypted2, _ := Decrypt(encrypted2, key)

	if string(decrypted1) != plaintext || string(decrypted2) != plaintext {
		t.Error("Decrypted values do not match original plaintext")
	}
}

// BenchmarkEncrypt benchmarks the Encrypt function.
func BenchmarkEncrypt(b *testing.B) {
	data := []byte("benchmark test data for encryption")
	key := key32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(data, key)
	}
}

// BenchmarkDecrypt benchmarks the Decrypt function.
func BenchmarkDecrypt(b *testing.B) {
	data := []byte("benchmark test data for encryption")
	key := key32
	encrypted, _ := Encrypt(data, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(encrypted, key)
	}
}

// BenchmarkEncryptToB64 benchmarks the EncryptToB64 function.
func BenchmarkEncryptToB64(b *testing.B) {
	data := "benchmark test data for encryption"
	key := key32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncryptToB64(data, key)
	}
}

// BenchmarkDecryptFromB64 benchmarks the DecryptFromB64 function.
func BenchmarkDecryptFromB64(b *testing.B) {
	data := "benchmark test data for encryption"
	key := key32
	encrypted, _ := EncryptToB64(data, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptFromB64(encrypted, key)
	}
}
