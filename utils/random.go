//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//

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
