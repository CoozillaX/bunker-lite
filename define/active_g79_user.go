package define

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bytes"
	"encoding/json"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ActiveG79User ..
type ActiveG79User struct {
	SessionID         string
	SessionStartTime  int64
	SessionExpireTime int64
	RecordG79UserData *g79.G79User
}

// EncodeActiveG79User ..
func EncodeActiveG79User(user ActiveG79User) []byte {
	buf := bytes.NewBuffer(nil)
	writer := protocol.NewWriter(buf, 0)

	writer.String(&user.SessionID)
	writer.Int64(&user.SessionStartTime)
	writer.Int64(&user.SessionExpireTime)

	writer.String(&user.RecordG79UserData.EntityID)
	writer.String(&user.RecordG79UserData.Account)
	writer.String(&user.RecordG79UserData.G79Token)
	writer.String(&user.RecordG79UserData.Sead)
	writer.String(&user.RecordG79UserData.Username)

	jsonBytes, _ := json.Marshal(user.RecordG79UserData.MpayUser)
	writer.ByteSlice(&jsonBytes)
	jsonBytes, _ = json.Marshal(user.RecordG79UserData.GameInfo)
	writer.ByteSlice(&jsonBytes)

	return buf.Bytes()
}

// DecodeActiveG79User ..
func DecodeActiveG79User(payload []byte) (user ActiveG79User) {
	var jsonBytes []byte

	user.RecordG79UserData = &g79.G79User{
		GameInfo: new(defines.GameInfo),
	}
	buf := bytes.NewBuffer(payload)
	reader := protocol.NewReader(buf, 0, false)

	reader.String(&user.SessionID)
	reader.Int64(&user.SessionStartTime)
	reader.Int64(&user.SessionExpireTime)

	reader.String(&user.RecordG79UserData.EntityID)
	reader.String(&user.RecordG79UserData.Account)
	reader.String(&user.RecordG79UserData.G79Token)
	reader.String(&user.RecordG79UserData.Sead)
	reader.String(&user.RecordG79UserData.Username)

	reader.ByteSlice(&jsonBytes)
	_ = json.Unmarshal(jsonBytes, &user.RecordG79UserData.MpayUser)
	reader.ByteSlice(&jsonBytes)
	_ = json.Unmarshal(jsonBytes, user.RecordG79UserData.GameInfo)

	return user
}
