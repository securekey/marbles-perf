/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	cryptorandom "crypto/rand"
	"encoding/base64"
	"math/big"
)

const alphaNumChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var alphaNumCharsLen = big.NewInt(int64(len(alphaNumChars)))

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := cryptorandom.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	if b, err := GenerateRandomBytes(s); err != nil {
		return "", err
	} else {
		return base64.RawURLEncoding.EncodeToString(b), nil
	}
}

// GenerateRandomAlphaNumericString generates random alpha numeric string
//
func GenerateRandomAlphaNumericString(size int) (string, error) {
	bytes := make([]byte, size)
	for i := range bytes {
		n, err := cryptorandom.Int(cryptorandom.Reader, alphaNumCharsLen)
		if err != nil {
			return "", err
		}
		bytes[i] = alphaNumChars[n.Int64()]
	}
	return string(bytes), nil
}
