package define

import (
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

//go:embed game_saves_encrypt.key
var keyBytes []byte
var GameSavesEncryptKey *rsa.PrivateKey

func init() {
	var err error
	keyBlock, _ := pem.Decode(keyBytes)
	GameSavesEncryptKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		panic(err)
	}
}

// GameSavesKey ..
type GameSavesKey struct {
	EulogistUserUniqueID string
	RentalServerNumber   string
	GameSavesAESCipher   []byte
}

func (g *GameSavesKey) MarshalKey(io protocol.IO) {
	io.String(&g.EulogistUserUniqueID)
	io.String(&g.RentalServerNumber)
}

func (g *GameSavesKey) MarshalData(io protocol.IO) {
	io.ByteSlice(&g.GameSavesAESCipher)
}
