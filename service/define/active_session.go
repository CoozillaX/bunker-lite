package define

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/mpay/android"
	"bytes"
	"encoding/json"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SessionExpireTimeSecond  = 2 * 60 * 60
	SessionMaxLifeTimeSecond = 12 * 60 * 60
)

const (
	SessionTypeMpayUser uint8 = iota
	SessionTypePeAuth
	SessionTypeSaAuth
)

// RecordG79User ..
type RecordG79User struct {
	InternalG79User *g79.G79User
	NextRefreshTime int64
}

// ActiveSession ..
type ActiveSession struct {
	SessionID         string
	SessionType       uint8
	SessionStartTime  int64
	SessionExpireTime int64
	RecordG79UserData *RecordG79User
}

func NewActiveSession() ActiveSession {
	return ActiveSession{
		RecordG79UserData: &RecordG79User{
			InternalG79User: &g79.G79User{
				MpayUser: &android.AndroidMpayUser{
					GameInfo: new(android.AndroidGameInfo),
				},
			},
		},
	}
}

// G79User ..
func (a *ActiveSession) G79User() *g79.G79User {
	return a.RecordG79UserData.InternalG79User
}

func (a *ActiveSession) MpayUser() *android.AndroidMpayUser {
	return a.RecordG79UserData.InternalG79User.MpayUser.(*android.AndroidMpayUser)
}

// EncodeActiveSession ..
func EncodeActiveSession(session ActiveSession) []byte {
	buf := bytes.NewBuffer(nil)
	writer := protocol.NewWriter(buf, 0)

	writer.String(&session.SessionID)
	writer.Uint8(&session.SessionType)
	writer.Int64(&session.SessionStartTime)
	writer.Int64(&session.SessionExpireTime)

	writer.Int64(&session.RecordG79UserData.NextRefreshTime)
	writer.String(&session.RecordG79UserData.InternalG79User.EntityID)
	writer.String(&session.RecordG79UserData.InternalG79User.Account)
	writer.String(&session.RecordG79UserData.InternalG79User.G79Token)
	writer.String(&session.RecordG79UserData.InternalG79User.Sead)
	writer.String(&session.RecordG79UserData.InternalG79User.Username)

	jsonBytes, _ := json.Marshal(session.MpayUser())
	writer.ByteSlice(&jsonBytes)
	jsonBytes, _ = json.Marshal(session.MpayUser().GameInfo)
	writer.ByteSlice(&jsonBytes)

	return buf.Bytes()
}

// DecodeActiveSession ..
func DecodeActiveSession(payload []byte) (session ActiveSession) {
	var jsonBytes []byte

	session = NewActiveSession()
	buf := bytes.NewBuffer(payload)
	reader := protocol.NewReader(buf, 0, false)

	reader.String(&session.SessionID)
	reader.Uint8(&session.SessionType)
	reader.Int64(&session.SessionStartTime)
	reader.Int64(&session.SessionExpireTime)

	reader.Int64(&session.RecordG79UserData.NextRefreshTime)
	reader.String(&session.RecordG79UserData.InternalG79User.EntityID)
	reader.String(&session.RecordG79UserData.InternalG79User.Account)
	reader.String(&session.RecordG79UserData.InternalG79User.G79Token)
	reader.String(&session.RecordG79UserData.InternalG79User.Sead)
	reader.String(&session.RecordG79UserData.InternalG79User.Username)

	reader.ByteSlice(&jsonBytes)
	_ = json.Unmarshal(jsonBytes, session.MpayUser())
	reader.ByteSlice(&jsonBytes)
	_ = json.Unmarshal(jsonBytes, session.MpayUser().GameInfo)

	return session
}
