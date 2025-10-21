package vitality_api

import (
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
)

//go:embed token_encrypt.key
var keyBytes []byte
var TokenEncryptKey *rsa.PrivateKey

func init() {
	var err error
	keyBlock, _ := pem.Decode(keyBytes)
	TokenEncryptKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		panic(err)
	}
}
