package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
)

const UserPassword = "qwertyQWERTY"

type Decoder struct {
	aesgcm cipher.AEAD
	nonce  []byte
}

func NewDecoder() *Decoder {
	key := sha256.Sum256([]byte(UserPassword))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil
	}
	nonce := key[len(key)-aesgcm.NonceSize():]

	return &Decoder{
		aesgcm: aesgcm,
		nonce:  nonce,
	}
}

func (d *Decoder) Encode(userID string) string {
	return hex.EncodeToString(d.aesgcm.Seal(nil, d.nonce, []byte(userID), nil))
}

func (d *Decoder) Decode(msg string) (string, error) {
	msgBytes, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	decoded, err := d.aesgcm.Open(nil, d.nonce, msgBytes, nil)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
