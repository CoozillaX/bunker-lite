package define

import (
	"bytes"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AuthServerAccount ..
type AuthServerAccount interface {
	AuthServerAddress() string
	AuthServerToken() string
	FormatInGame() string
	IsStdAccount() bool
	UpdateData(newData map[string]any)
}

// EncodeAuthServerAccount ..
func EncodeAuthServerAccount(account AuthServerAccount) []byte {
	buf := bytes.NewBuffer(nil)
	writer := protocol.NewWriter(buf, 0)

	isStdAccount := account.IsStdAccount()
	writer.Bool(&isStdAccount)

	if isStdAccount {
		stdAccount := account.(*StdAuthServerAccount)
		writer.String(&stdAccount.gameNickName)
		writer.String(&stdAccount.authServerToken)
		return buf.Bytes()
	}

	customAccount := account.(*CustomAuthServerAccount)
	writer.Uint8(&customAccount.internalAccountID)
	writer.String(&customAccount.authServerAddress)
	writer.String(&customAccount.authServerToken)
	return buf.Bytes()
}

// DecodeAuthServerAccount ..
func DecodeAuthServerAccount(payload []byte) AuthServerAccount {
	buf := bytes.NewBuffer(payload)
	reader := protocol.NewReader(buf, 0, false)

	isStdAccount := false
	reader.Bool(&isStdAccount)

	if isStdAccount {
		account := StdAuthServerAccount{}
		reader.String(&account.gameNickName)
		reader.String(&account.authServerToken)
		return &account
	}

	account := CustomAuthServerAccount{}
	reader.Uint8(&account.internalAccountID)
	reader.String(&account.authServerAddress)
	reader.String(&account.authServerToken)
	return &account
}

// StdAuthServerAccount ..
type StdAuthServerAccount struct {
	gameNickName    string
	g79UserUID      string
	authServerToken string
}

func (s *StdAuthServerAccount) IsStdAccount() bool {
	return true
}

func (s *StdAuthServerAccount) FormatInGame() string {
	return fmt.Sprintf("§r§l§e%s §r§7(§fUID §7- §b%s§7)", s.gameNickName, s.g79UserUID)
}

func (s *StdAuthServerAccount) AuthServerAddress() string {
	return StdAuthServerAddress
}

func (s *StdAuthServerAccount) AuthServerToken() string {
	return s.authServerToken
}

func (s *StdAuthServerAccount) UpdateData(newData map[string]any) {
	*s = StdAuthServerAccount{
		gameNickName:    newData["gameNickName"].(string),
		g79UserUID:      newData["g79UserUID"].(string),
		authServerToken: newData["authServerToken"].(string),
	}
}

// CustomAuthServerAccount ..
type CustomAuthServerAccount struct {
	internalAccountID uint8
	authServerAddress string
	authServerToken   string
}

func (c *CustomAuthServerAccount) IsStdAccount() bool {
	return false
}

func (c *CustomAuthServerAccount) FormatInGame() string {
	return fmt.Sprintf("§r§l§e账户 ID §f- §b%d", c.internalAccountID)
}

func (c *CustomAuthServerAccount) AuthServerAddress() string {
	return c.authServerAddress
}

func (c *CustomAuthServerAccount) AuthServerToken() string {
	return c.authServerToken
}

func (c *CustomAuthServerAccount) UpdateData(newData map[string]any) {
	*c = CustomAuthServerAccount{
		internalAccountID: newData["internalAccountID"].(uint8),
		authServerAddress: newData["authServerAddress"].(string),
		authServerToken:   newData["authServerToken"].(string),
	}
}
