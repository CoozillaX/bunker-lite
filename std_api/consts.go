package std_api

import (
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
)

const (
	session_key_entity_id      = "SESSION_KEY_ENTITY_ID"
	session_key_engine_version = "SESSION_KEY_ENGINE_VERSION"
	session_key_patch_version  = "SESSION_KEY_PATCH_VERSION"
)

//go:embed phoenix_login.key
var keyBytes []byte
var PhoenixLoginKey *rsa.PublicKey

func init() {
	var err error
	keyBlock, _ := pem.Decode(keyBytes)
	PhoenixLoginKey, err = x509.ParsePKCS1PublicKey(keyBlock.Bytes)
	if err != nil {
		panic(err)
	}
}
