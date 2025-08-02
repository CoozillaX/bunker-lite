package define

import (
	"bytes"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// EulogistUser ..
type EulogistUser struct {
	UserUniqueID        string
	UserName            string
	UserPermissionLevel uint8

	UserPasswordSum256 []byte
	EulogistToken      string
	UnbanUnixTime      int64

	MultipleAuthServerAccounts []AuthServerAccount
	RentalServerConfig         []RentalServerConfig
	RentalServerCanManage      []string

	CurrentAuthServerAccount   protocol.Optional[AuthServerAccount]
	ProvidedPeAuthData         string
	DisableGlobalOpertorVerify bool
	CanAccessAnyRentalServer   bool
	CanGetGameSavesKeyCipher   bool
	CanGetHelperToken          bool
}

// EncodeEulogistUser ..
func EncodeEulogistUser(user EulogistUser) []byte {
	buf := bytes.NewBuffer(nil)
	writer := protocol.NewWriter(buf, 0)

	writer.String(&user.UserUniqueID)
	writer.String(&user.UserName)
	writer.Uint8(&user.UserPermissionLevel)
	writer.ByteSlice(&user.UserPasswordSum256)
	writer.String(&user.EulogistToken)
	writer.Varint64(&user.UnbanUnixTime)
	writer.String(&user.ProvidedPeAuthData)
	writer.Bool(&user.DisableGlobalOpertorVerify)
	writer.Bool(&user.CanAccessAnyRentalServer)
	writer.Bool(&user.CanGetGameSavesKeyCipher)
	writer.Bool(&user.CanGetHelperToken)
	protocol.SliceUint8Length(writer, &user.RentalServerConfig)
	protocol.FuncSliceUint16Length(writer, &user.RentalServerCanManage, writer.String)

	account, ok := user.CurrentAuthServerAccount.Value()
	writer.Bool(&ok)
	if ok {
		accountBytes := EncodeAuthServerAccount(account)
		writer.ByteSlice(&accountBytes)
	}

	slicenLen := uint8(len(user.MultipleAuthServerAccounts))
	writer.Uint8(&slicenLen)
	for _, account := range user.MultipleAuthServerAccounts {
		accountBytes := EncodeAuthServerAccount(account)
		writer.ByteSlice(&accountBytes)
	}

	return buf.Bytes()
}

// EncodeEulogistUser ..
func DecodeEulogistUser(payload []byte) (user EulogistUser) {
	var accountBytes []byte
	var haveCurrentAuthServerAccount bool
	var slicenLen uint8

	buf := bytes.NewBuffer(payload)
	reader := protocol.NewReader(buf, 0, false)

	reader.String(&user.UserUniqueID)
	reader.String(&user.UserName)
	reader.Uint8(&user.UserPermissionLevel)
	reader.ByteSlice(&user.UserPasswordSum256)
	reader.String(&user.EulogistToken)
	reader.Varint64(&user.UnbanUnixTime)
	reader.String(&user.ProvidedPeAuthData)
	reader.Bool(&user.DisableGlobalOpertorVerify)
	reader.Bool(&user.CanAccessAnyRentalServer)
	reader.Bool(&user.CanGetGameSavesKeyCipher)
	reader.Bool(&user.CanGetHelperToken)
	protocol.SliceUint8Length(reader, &user.RentalServerConfig)
	protocol.FuncSliceUint16Length(reader, &user.RentalServerCanManage, reader.String)

	reader.Bool(&haveCurrentAuthServerAccount)
	if haveCurrentAuthServerAccount {
		reader.ByteSlice(&accountBytes)
		account := DecodeAuthServerAccount(accountBytes)
		user.CurrentAuthServerAccount = protocol.Option(account)
	}

	reader.Uint8(&slicenLen)
	for range slicenLen {
		reader.ByteSlice(&accountBytes)
		user.MultipleAuthServerAccounts = append(
			user.MultipleAuthServerAccounts,
			DecodeAuthServerAccount(accountBytes),
		)
	}

	return
}
