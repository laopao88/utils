package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
)

func RSABase64Encrypt(data []byte, publicKey string) (string, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %v", err)
	}

	pubInterface, err := x509.ParsePKIXPublicKey(decodedPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}
	rsaPublicKey := pubInterface.(*rsa.PublicKey)
	blockLength := rsaPublicKey.N.BitLen()/8 - 11
	if len(data) <= blockLength {
		v15, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data)
		if err != nil {
			return "", err
		}
		return string(v15), nil
	}
	buffer := bytes.NewBufferString("")
	pages := len(data) / blockLength
	for index := 0; index <= pages; index++ {
		start := index * blockLength
		end := (index + 1) * blockLength
		if index == pages {
			if start == len(data) {
				continue
			}
			end = len(data)
		}
		chunk, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data[start:end])
		if err != nil {
			return "", err
		}
		buffer.Write(chunk)
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func RSABase64Decrypt(encodedData string, publicKey string) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %v", err)
	}
	pubInterface, err := x509.ParsePKIXPublicKey(decodedPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}
	rsaPublicKey := pubInterface.(*rsa.PublicKey)
	var result string
	for len(decodedData) > 0 {
		decodePart := decodedData[:128]
		plain := RsaPublicDecrypt(rsaPublicKey, decodePart)
		result += string(plain)
		decodedData = decodedData[128:]
	}
	return result, nil
}

func RsaPublicDecrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}
