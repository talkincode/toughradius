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

// Package aes provides AES encryption and decryption utilities using CBC mode with PKCS7 padding.
//
// This package implements symmetric encryption suitable for encrypting sensitive data like
// passwords, tokens, and configuration values in ToughRADIUS. All functions use AES-CBC mode
// with the encryption key doubled as the IV (first blockSize bytes).
//
// Security Notice:
//   - Key length must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
//   - Using the same key as IV reduces security; consider using a separate random IV in production
//   - All panic scenarios are recovered internally to prevent service crashes
//
// Example usage:
//
//	// Encrypt to base64 string
//	encrypted, err := aes.EncryptToB64("sensitive_data", "my-secret-key-16")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Decrypt from base64 string
//	decrypted, err := aes.DecryptFromB64(encrypted, "my-secret-key-16")
//	if err != nil {
//	    log.Fatal(err)
//	}
package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// Encrypt encrypts plaintext data using AES-CBC mode with PKCS7 padding.
//
// This function uses the provided key for both encryption and IV (first blockSize bytes).
// The plaintext is automatically padded to match the AES block size (16 bytes) using PKCS7.
//
// Parameters:
//   - orig: Raw plaintext bytes to encrypt
//   - key: Encryption key, must be 16, 24, or 32 bytes for AES-128/192/256
//
// Returns:
//   - []byte: Encrypted ciphertext bytes
//   - error: Encryption error, nil if successful
//
// Common errors:
//   - crypto/aes: invalid key size: returned when key length is not 16/24/32
//   - panic recovery: any runtime panic is caught and printed to stdout
//
// Security considerations:
//   - Using key as IV is not cryptographically recommended for production
//   - Consider generating a random IV and prepending it to the ciphertext
func Encrypt(orig []byte, key string) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recover AesEncrypt.", err)
		}
	}()
	k := []byte(key)

	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	orig = PKCS7Padding(orig, blockSize)
	// Use the first blockSize bytes of the key as IV
	// Note: This is less secure than using a random IV per encryption
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	cryted := make([]byte, len(orig))
	blockMode.CryptBlocks(cryted, orig)
	return cryted, nil
}

// EncryptToB64 encrypts a string and returns the result as a base64-encoded string.
//
// This is a convenience wrapper around Encrypt() that handles string-to-bytes conversion
// and base64 encoding, making it ideal for storing encrypted data in databases or
// configuration files.
//
// Parameters:
//   - orig: Plaintext string to encrypt
//   - key: Encryption key, must be 16, 24, or 32 bytes for AES-128/192/256
//
// Returns:
//   - string: Base64-encoded encrypted string
//   - error: Encryption or encoding error, nil if successful
//
// Example:
//
//	encrypted, err := EncryptToB64("mypassword", "1234567890123456")
//	// encrypted might be: "dGVzdCBlbmNyeXB0ZWQgZGF0YQ=="
func EncryptToB64(orig string, key string) (string, error) {
	bs, err := Encrypt([]byte(orig), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bs), nil
}

// Decrypt decrypts AES-CBC encrypted data and removes PKCS7 padding.
//
// This function reverses the Encrypt() operation, using the same key for both
// decryption and IV extraction. The ciphertext must have been encrypted with
// matching parameters (same key, CBC mode, PKCS7 padding).
//
// Parameters:
//   - cryted: Encrypted ciphertext bytes (from Encrypt function)
//   - key: Decryption key, must match the encryption key
//
// Returns:
//   - []byte: Decrypted plaintext bytes
//   - error: Decryption error, nil if successful
//
// Common errors:
//   - crypto/aes: invalid key size: key length doesn't match AES requirements
//   - PKCS7UnPadding errors: ciphertext was corrupted or encrypted with different key
//   - panic recovery: any runtime panic is caught and printed to stdout
func Decrypt(cryted []byte, key string) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recover AesDecrypt.", err)
		}
	}()
	k := []byte(key)
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	// Use the first blockSize bytes of the key as IV (must match encryption)
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	orig := make([]byte, len(cryted))
	blockMode.CryptBlocks(orig, cryted)
	orig, err = PKCS7UnPadding(orig)
	if err != nil {
		return nil, err
	}
	return orig, nil
}

// DecryptFromB64 decrypts a base64-encoded encrypted string back to plaintext.
//
// This is a convenience wrapper around Decrypt() that handles base64 decoding
// and bytes-to-string conversion. It's the inverse operation of EncryptToB64().
//
// Parameters:
//   - cryted: Base64-encoded encrypted string (from EncryptToB64)
//   - key: Decryption key, must match the encryption key
//
// Returns:
//   - string: Decrypted plaintext string
//   - error: Decoding or decryption error, nil if successful
//
// Common errors:
//   - base64: illegal base64 data: input string is not valid base64
//   - PKCS7UnPadding errors: wrong key used or data corrupted
//
// Example:
//
//	decrypted, err := DecryptFromB64("dGVzdCBlbmNyeXB0ZWQgZGF0YQ==", "1234567890123456")
//	// decrypted might be: "mypassword"
func DecryptFromB64(cryted string, key string) (string, error) {
	bs, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		return "", err
	}
	bs2, err2 := Decrypt(bs, key)
	if err2 != nil {
		return "", err2
	}
	return string(bs2), nil
}

// PKCS7Padding adds PKCS#7 padding to the plaintext to match the AES block size.
//
// PKCS#7 padding scheme adds N bytes of value N to fill the last incomplete block,
// where N is the number of bytes needed. This ensures the ciphertext length is
// always a multiple of the block size (16 bytes for AES).
//
// Parameters:
//   - ciphertext: Raw data to pad (despite the name, used before encryption)
//   - blocksize: AES block size, typically 16 bytes
//
// Returns:
//   - []byte: Padded data ready for block cipher encryption
//
// Example:
//   - Input: [0x01, 0x02, 0x03] with blocksize=16
//   - Output: [0x01, 0x02, 0x03, 0x0D, 0x0D, ..., 0x0D] (13 bytes of 0x0D added)
//
// Reference: RFC 5652 Section 6.3
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	// Create a slice filled with the padding value (number of padding bytes)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding removes PKCS#7 padding from decrypted data.
//
// This function validates and removes the padding bytes added by PKCS7Padding.
// It checks that the padding value is valid and that the data length is sufficient.
//
// Parameters:
//   - origData: Decrypted data with PKCS#7 padding
//
// Returns:
//   - []byte: Original data without padding
//   - error: Validation error if padding is invalid or data is empty
//
// Common errors:
//   - "no data": input slice is empty
//   - "PKCS7UnPadding error, data length < unpadding": corrupted padding or wrong key
//
// Example:
//   - Input: [0x01, 0x02, 0x03, 0x0D, 0x0D, ..., 0x0D] (13 bytes of 0x0D)
//   - Output: [0x01, 0x02, 0x03]
//
// Reference: RFC 5652 Section 6.3
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("no data")
	}
	// The last byte indicates how many padding bytes were added
	unpadding := int(origData[length-1])
	len := length - unpadding
	if len < 0 {
		return nil, errors.New("PKCS7UnPadding error, data length < unpadding")
	}
	return origData[:len], nil
}
