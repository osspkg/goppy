/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import "go.osspkg.com/goppy/v3/auth/token/algorithm"

//go:generate easyjson

//easyjson:json
type Header struct {
	//Algorithm to verify the signature on the token
	Algorithm algorithm.Name `json:"alg"`
	//Token type
	TokenType Type `json:"typ"`
	//Case-sensitive unique identifier of the token even among different issuers
	TokenID string `json:"jti,omitempty"`
	//Identifies principal that issued the JWT
	Issuer string `json:"iss,omitempty"`
	//A hint indicating which key the client used to generate the token signature
	KeyID string `json:"kid,omitempty"`
	//Identifies the recipients that the JWT is intended for
	Audience string `json:"aud,omitempty"`
	//Identifies the time at which the JWT was issued
	IssuedAt int64 `json:"iat"`
	//Identifies the expiration time on and after which the JWT must not be accepted for processing
	ExpiresAt int64 `json:"exp"`
}
