package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// AllowListConfig ..
type AllowListConfig struct {
	EulogistUserUniqueID       string
	DisableGlobalOpertorVerify bool
	CanGetGameSavesKeyCipher   bool
}

func (a *AllowListConfig) Marshal(io protocol.IO) {
	io.String(&a.EulogistUserUniqueID)
	io.Bool(&a.DisableGlobalOpertorVerify)
	io.Bool(&a.CanGetGameSavesKeyCipher)
}
