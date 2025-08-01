package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// AuthHelperAccount ..
type AuthHelperAccount struct {
	AccountUniqueID string
	AccountToken    string
	GameNickName    string
	MpayUserData    string
}

func (a *AuthHelperAccount) Marshal(io protocol.IO) {
	io.String(&a.AccountUniqueID)
	io.String(&a.AccountToken)
	io.String(&a.GameNickName)
	io.String(&a.MpayUserData)
}
