// Copyright 2016 The Gogs Authors. All rights reserved.
// Copyright 2016 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package generate

import (
	"encoding/base64"
	"fmt"

	"forgejo.org/modules/util"
)

const defaultJwtSecretLen = 32

// DecodeJwtSecret decodes a base64 encoded jwt secret into bytes, and check its length
func DecodeJwtSecret(src string) ([]byte, error) {
	encoding := base64.RawURLEncoding
	decoded := make([]byte, encoding.DecodedLen(len(src))+3)
	if n, err := encoding.Decode(decoded, []byte(src)); err != nil {
		return nil, fmt.Errorf("JwtSecret decode failed: %v", err)
	} else if n != defaultJwtSecretLen {
		return nil, fmt.Errorf("invalid base64 decoded length: %d, expects: %d", n, defaultJwtSecretLen)
	}
	return decoded[:defaultJwtSecretLen], nil
}

// NewJwtSecret generates a new base64 encoded value intended to be used for JWT secrets.
func NewJwtSecret() ([]byte, string) {
	bytes := util.CryptoRandomBytes(32)
	return bytes, base64.RawURLEncoding.EncodeToString(bytes)
}

// NewSecretKey generate a new value intended to be used by SECRET_KEY.
func NewSecretKey() string {
	return util.CryptoRandomString(util.RandomStringHigh)
}
