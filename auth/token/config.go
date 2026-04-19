/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v3/auth/token/algorithm"
)

const (
	SourceFile = "file"
	SourceRaw  = "raw"
	SourceEnv  = "env"

	FormatRaw    = "raw"
	FormatBase64 = "base64"
	FormatHex    = "hex"
)

type (
	ConfigGroup struct {
		JWT Config `yaml:"jwt"`
	}

	Config struct {
		Option ConfigOption `yaml:"option"`
		Sign   ConfigSign   `yaml:"sign"`
		//JWKS   []ConfigJWKS `yaml:"jwks"`
	}

	ConfigSign struct {
		Issuer string      `yaml:"issuer"`
		Type   Type        `yaml:"type"`
		Keys   []ConfigKey `yaml:"keys"`
	}

	ConfigKey struct {
		ID   string         `yaml:"id"`
		Algo algorithm.Name `yaml:"algo"`
		Key  KeyLink        `yaml:"key"`
		Cert KeyLink        `yaml:"cert,omitempty"`
	}

	ConfigJWKS struct {
		Issuer   string            `yaml:"issuer"`
		Uri      string            `yaml:"uri"`
		Headers  map[string]string `yaml:"headers,omitempty"`
		Interval time.Duration     `yaml:"interval"`
	}

	ConfigOption struct {
		HeaderName string `yaml:"header_name,omitempty"`
		CookieName string `yaml:"cookie_name,omitempty"`
		SecureOnly bool   `yaml:"secure_only,omitempty"`
		Audience   string `yaml:"audience"`
	}
)

type (
	KeyLink    string
	SourceData struct {
		Source string
		Format string
	}
)

func (v KeyLink) getSource() (SourceData, error) {
	index := strings.Index(string(v), ":")
	if index == -1 {
		return SourceData{}, fmt.Errorf("invalid key link: %s", v)
	}
	sources := strings.Split(string(v[:index]), ",")
	if len(sources) != 2 {
		return SourceData{}, fmt.Errorf("invalid key link: %s", v)
	}
	return SourceData{
		Source: strings.TrimSpace(strings.ToLower(sources[0])),
		Format: strings.TrimSpace(strings.ToLower(sources[1])),
	}, nil
}

func (v KeyLink) getValue() string {
	index := strings.Index(string(v), ":")
	if index == -1 {
		return ""
	}
	if len(v)-1 < index+1 {
		return ""
	}
	return string(v[index+1:])
}

func (v KeyLink) getBytes() ([]byte, error) {
	var result []byte

	src, err := v.getSource()
	if err != nil {
		return nil, err
	}

	val := v.getValue()

	switch src.Source {
	case SourceRaw:
		result = []byte(val)
	case SourceEnv:
		data, ok := os.LookupEnv(val)
		if !ok {
			return nil, fmt.Errorf("environment `%s` not found", val)
		}
		result = []byte(data)
	case SourceFile:
		if !fs.FileExist(val) {
			return result, fmt.Errorf("file `%s` not found", val)
		}
		if result, err = os.ReadFile(val); err != nil {
			return nil, fmt.Errorf("failed to read file `%s`: %w", val, err)
		}
	default:
		return nil, fmt.Errorf("unsupported source `%s`: %s", src.Source, v)
	}

	switch src.Format {
	case FormatRaw:
	case FormatBase64:
		b, err := base64.StdEncoding.AppendDecode(nil, result)
		if err != nil {
			return nil, fmt.Errorf("failed to base64 decode `%s`: %w", val, err)
		}
		result = b
	case FormatHex:
		b, err := hex.AppendDecode(nil, result)
		if err != nil {
			return nil, fmt.Errorf("failed to base64 decode `%s`: %w", val, err)
		}
		result = b
	default:
		return nil, fmt.Errorf("unsupported format `%s`: %s", src.Format, v)
	}

	return result, nil
}

func (v ConfigKey) Validate() error {
	if len(v.ID) == 0 {
		return fmt.Errorf("id is required")
	}
	if _, err := algorithm.Get(v.Algo); err != nil {
		return fmt.Errorf("algo `%s` is invalid: %w", v.Algo, err)
	}

	if len(v.Key) == 0 {
		return fmt.Errorf("key is required")
	}
	if len(v.Cert) == 0 {
		return fmt.Errorf("cert is required")
	}

	if keyValue := v.Key.getValue(); len(keyValue) < 1 {
		return fmt.Errorf("key is required")
	}
	if certValue := v.Cert.getValue(); len(certValue) < 1 {
		return fmt.Errorf("cert is required")
	}

	keySources, err := v.Key.getSource()
	if err != nil {
		return fmt.Errorf("key must have source")
	}

	certSources, err := v.Cert.getSource()
	if err != nil {
		return fmt.Errorf("cert must have source")
	}

	switch keySources.Source {
	case SourceFile, SourceEnv, SourceRaw:
	default:
		return fmt.Errorf(
			"invalid key source: %s (use %s)",
			keySources.Source,
			strings.Join([]string{SourceFile, SourceEnv, SourceRaw}, ","),
		)
	}

	switch keySources.Format {
	case FormatBase64, FormatHex, FormatRaw:
	default:
		return fmt.Errorf(
			"invalid key format: %s (use %s)",
			keySources.Source,
			strings.Join([]string{FormatBase64, FormatHex, FormatRaw}, ","),
		)
	}

	switch certSources.Source {
	case SourceFile, SourceEnv, SourceRaw:
	default:
		return fmt.Errorf(
			"invalid cert source: %s (use %s)",
			certSources.Source,
			strings.Join([]string{SourceFile, SourceEnv, SourceRaw}, ","),
		)
	}

	switch certSources.Format {
	case FormatBase64, FormatHex, FormatRaw:
	default:
		return fmt.Errorf(
			"invalid cert format: %s (use %s)",
			certSources.Source,
			strings.Join([]string{FormatBase64, FormatHex, FormatRaw}, ","),
		)
	}

	return nil
}

func (v *ConfigGroup) Default() error {
	if v == nil {
		return fmt.Errorf("jwt config: group config is nil")
	}

	v.JWT.Option = ConfigOption{
		HeaderName: "Authorization",
		CookieName: "jwt",
		SecureOnly: true,
		Audience:   "example:platform",
	}

	v.JWT.Sign = ConfigSign{
		Issuer: "example:service",
		Type:   TypeJWT,
	}

	//v.JWT.JWKS = append(v.JWT.JWKS, ConfigJWKS{
	//	Issuer:   "example:id",
	//	Uri:      "https://id.example.com/.well-known/jwks.json",
	//	Interval: time.Hour,
	//	Headers: map[string]string{
	//		"X-Auth-Id": "example",
	//	},
	//})

	name := algorithm.EdDSA
	algObj, err := algorithm.Get(name)
	if err != nil {
		return fmt.Errorf("jwt config: algorithm '%s': %w", name, err)
	}

	for i := 0; i < 10; i++ {
		keyId := uuid.NewString()
		keyAny, err := algObj.GenerateKeys()
		if err != nil {
			return fmt.Errorf("jwt config: generate keys for algorithm '%s': %w", name, err)
		}
		keyStr, err := algObj.EncodeBase64(keyAny)
		if err != nil {
			return fmt.Errorf("jwt config: encode key for algorithm '%s': %w", name, err)
		}

		v.JWT.Sign.Keys = append(v.JWT.Sign.Keys, ConfigKey{
			ID:   keyId,
			Algo: name,
			Key:  KeyLink(SourceRaw + "," + FormatBase64 + ":" + keyStr.Private),
			Cert: KeyLink(SourceRaw + "," + FormatBase64 + ":" + keyStr.Public),
		})
	}

	return nil
}

func (v Config) Validate() error {
	if len(v.Sign.Keys) == 0 {
		return fmt.Errorf("jwt config: keys config is empty")
	}

	if v.Sign.Issuer == "" {
		return fmt.Errorf("jwt config: sign issuer is empty")
	}

	switch v.Sign.Type {
	case TypeJWT:
	default:
		return fmt.Errorf("jwt config: sign token_type[%s] is invalid", v.Sign.Type)
	}

	for i, k := range v.Sign.Keys {
		if err := k.Validate(); err != nil {
			return fmt.Errorf("jwt config: sign keys[%d] is invalid: %w", i, err)
		}
	}

	//for i, jwk := range v.JWKS {
	//	if jwk.Issuer == "" {
	//		return fmt.Errorf("jwt config: jwks[%d] issuer is empty", i)
	//	}
	//
	//	if jwk.Uri == "" {
	//		return fmt.Errorf("jwt config: jwks[%d] uri is empty", i)
	//	}
	//}

	return nil
}

func (v *ConfigGroup) Validate() error {
	if v == nil {
		return fmt.Errorf("jwt config: group config is nil")
	}

	if err := v.JWT.Validate(); err != nil {
		return fmt.Errorf("jwt config: %w", err)
	}

	return nil
}
