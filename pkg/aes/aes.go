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
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

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
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	cryted := make([]byte, len(orig))
	blockMode.CryptBlocks(cryted, orig)
	return cryted, nil
}

func EncryptToB64(orig string, key string) (string, error) {
	bs, err := Encrypt([]byte(orig), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bs), nil
}

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
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	orig := make([]byte, len(cryted))
	blockMode.CryptBlocks(orig, cryted)
	orig, err = PKCS7UnPadding(orig)
	if err != nil {
		return nil, err
	}
	return orig, nil
}

func DecryptFromB64(cryted string, key string) (string, error) {
	bs, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		return "", err
	}
	bs2, err2 := Decrypt(bs, key)
	if err2 != nil {
		return "", err
	}
	return string(bs2), nil
}

func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("no data")
	}
	unpadding := int(origData[length-1])
	len := length - unpadding
	if len < 0 {
		return nil, errors.New("PKCS7UnPadding error, data length < unpadding")
	}
	return origData[:len], nil
}
