package des

import (
	"bytes"
	"crypto/des"
	"encoding/base64"
	"errors"
)

// Plaintext complement algorithm
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// Plaintext subtract algorithm
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}
func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

func DesEncrypt(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	// Complement the plaintext data
	src = PKCS5Padding(src, bs)
	if len(src)%bs != 0 {
		return nil, errors.New("Need a multiple of the blocksize")
	}
	out := make([]byte, len(src))
	dst := out
	// Encrypt the plaintext in block-sized chunks
	// Use goroutines for parallel encryption if needed
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	dstBase64 := make([]byte, base64.StdEncoding.EncodedLen(len(out)))
	base64.StdEncoding.Encode(dstBase64, out)
	return dstBase64, nil
}

func DesDecrypt(src, key []byte) ([]byte, error) {
	srcBase64 := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(srcBase64, src)
	if err != nil {
		return nil, err
	}
	srcUnBase64 := srcBase64[:n]
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(srcUnBase64))
	dst := out
	bs := block.BlockSize()
	if len(srcUnBase64)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}
	for len(srcUnBase64) > 0 {
		block.Decrypt(dst, srcUnBase64[:bs])
		srcUnBase64 = srcUnBase64[bs:]
		dst = dst[bs:]
	}
	out = PKCS5UnPadding(out)
	return out, nil
}
