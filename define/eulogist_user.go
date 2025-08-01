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

	UserPasswordSum512 string
	EulogistToken      string

	MultipleAuthServerAccounts []AuthServerAccount
	RentalServerConfig         []RentalServerConfig
	RentalServerCanManage      []string

	CurrentAuthServerAccount    AuthServerAccount
	AccessRentalServerWithoutOP bool
}

// EncodeEulogistUser ..
func EncodeEulogistUser(user EulogistUser) []byte {
	buf := bytes.NewBuffer(nil)
	writer := protocol.NewWriter(buf, 0)

	writer.String(&user.UserUniqueID)
	writer.String(&user.UserName)
	writer.Uint8(&user.UserPermissionLevel)
	writer.String(&user.UserPasswordSum512)
	writer.String(&user.EulogistToken)
	writer.Bool(&user.AccessRentalServerWithoutOP)
	protocol.SliceUint8Length(writer, &user.RentalServerConfig)
	protocol.FuncSliceUint16Length(writer, &user.RentalServerCanManage, writer.String)

	accountBytes := EncodeAuthServerAccount(user.CurrentAuthServerAccount)
	writer.ByteSlice(&accountBytes)

	slicenLen := uint8(len(user.MultipleAuthServerAccounts))
	writer.Uint8(&slicenLen)
	for _, account := range user.MultipleAuthServerAccounts {
		accountBytes = EncodeAuthServerAccount(account)
		writer.ByteSlice(&accountBytes)
	}

	return buf.Bytes()
}

// EncodeEulogistUser ..
func DecodeEulogistUser(payload []byte) (user EulogistUser) {
	var accountBytes []byte
	var slicenLen uint8

	buf := bytes.NewBuffer(payload)
	reader := protocol.NewReader(buf, 0, false)

	reader.String(&user.UserUniqueID)
	reader.String(&user.UserName)
	reader.Uint8(&user.UserPermissionLevel)
	reader.String(&user.UserPasswordSum512)
	reader.String(&user.EulogistToken)
	reader.Bool(&user.AccessRentalServerWithoutOP)
	protocol.SliceUint8Length(reader, &user.RentalServerConfig)
	protocol.FuncSliceUint16Length(reader, &user.RentalServerCanManage, reader.String)

	reader.ByteSlice(&accountBytes)
	user.CurrentAuthServerAccount = DecodeAuthServerAccount(accountBytes)

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
