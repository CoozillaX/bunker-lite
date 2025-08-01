package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// AuthServerHelper ..
type AuthServerHelper struct {
	HelperUniqueID string
	HelperToken    string
	GameNickName   string
	G79UserUID     string
	MpayUserData   []byte
}

func (a *AuthServerHelper) Marshal(io protocol.IO) {
	io.String(&a.HelperUniqueID)
	io.String(&a.HelperToken)
	io.String(&a.GameNickName)
	io.String(&a.G79UserUID)
	io.ByteSlice(&a.MpayUserData)
}
