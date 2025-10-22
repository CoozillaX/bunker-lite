package define

import (
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
)

const (
	AddressPhoenixAPI  = "https://yorha.eulogist-api.icu"
	AddressEulogistAPI = "https://yorha.eulogist-api.icu/eulogist_api"
	AddressVitalityAPI = "https://yorha.eulogist-api.icu/vitality_api"
	UserPasswordSalt   = "YoRHa"
)

const (
	AuthServerAccountTypeStd uint8 = iota
	AuthServerAccountTypeCustom
)

const (
	UserPermissionSystem = iota
	UserPermissionAdmin
	UserPermissionAdvance
	UserPermissionNormal
	UserPermissionNone
	UserPermissionDefault = UserPermissionNormal
)

//go:embed game_saves_encrypt.key
var gameSavesKeyBytes []byte
var GameSavesEncryptKey *rsa.PrivateKey

//go:embed token_encrypt.key
var tokenEncryptKeyBytes []byte
var TokenEncryptKey *rsa.PrivateKey

func init() {
	var err error

	keyBlock, _ := pem.Decode(gameSavesKeyBytes)
	GameSavesEncryptKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		panic(err)
	}

	keyBlock, _ = pem.Decode(tokenEncryptKeyBytes)
	TokenEncryptKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		panic(err)
	}
}
