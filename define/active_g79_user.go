package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ActiveG79User ..
type ActiveG79User struct {
	SessionID         string
	G79UserData       []byte
	G79UserExpireTime int64
}

func (a *ActiveG79User) Marshal(io protocol.IO) {
	io.String(&a.SessionID)
	io.ByteSlice(&a.G79UserData)
	io.Int64(&a.G79UserExpireTime)
}
