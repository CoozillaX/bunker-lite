package define

import (
	"bytes"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AuthServerAccount ..
type AuthServerAccount interface {
	AuthServerAddress() string
	AuthServerSecret() string
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
		writer.String(&stdAccount.authHelperUniqueID)
		return buf.Bytes()
	}

	customAccount := account.(*CustomAuthServerAccount)
	writer.Varuint32(&customAccount.internalAccountID)
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
		reader.String(&account.authHelperUniqueID)
		return &account
	}

	account := CustomAuthServerAccount{}
	reader.Varuint32(&account.internalAccountID)
	reader.String(&account.authServerAddress)
	reader.String(&account.authServerToken)
	return &account
}

// StdAuthServerAccount ..
type StdAuthServerAccount struct {
	gameNickName       string
	g79UserUID         string
	authHelperUniqueID string
}

func (s *StdAuthServerAccount) IsStdAccount() bool {
	return true
}

func (s *StdAuthServerAccount) FormatInGame() string {
	return fmt.Sprintf("§r§l§e%s §r§7(§fUID §7- §b%s§7)§r", s.gameNickName, s.g79UserUID)
}

func (s *StdAuthServerAccount) AuthServerAddress() string {
	return StdAuthServerAddress
}

func (s *StdAuthServerAccount) AuthServerSecret() string {
	return s.authHelperUniqueID
}

func (s *StdAuthServerAccount) G79UserUID() string {
	return s.g79UserUID
}

func (s *StdAuthServerAccount) UpdateData(newData map[string]any) {
	*s = StdAuthServerAccount{
		gameNickName:       newData["gameNickName"].(string),
		g79UserUID:         newData["g79UserUID"].(string),
		authHelperUniqueID: newData["authHelperUniqueID"].(string),
	}
}

// CustomAuthServerAccount ..
type CustomAuthServerAccount struct {
	internalAccountID uint32
	authServerAddress string
	authServerToken   string
}

func (c *CustomAuthServerAccount) IsStdAccount() bool {
	return false
}

func (c *CustomAuthServerAccount) FormatInGame() string {
	return fmt.Sprintf("§r§l§e账户 ID §f- §b%d§r", c.internalAccountID)
}

func (c *CustomAuthServerAccount) AuthServerAddress() string {
	return c.authServerAddress
}

func (c *CustomAuthServerAccount) AuthServerSecret() string {
	return c.authServerToken
}

func (c *CustomAuthServerAccount) UpdateData(newData map[string]any) {
	*c = CustomAuthServerAccount{
		internalAccountID: newData["internalAccountID"].(uint32),
		authServerAddress: newData["authServerAddress"].(string),
		authServerToken:   newData["authServerToken"].(string),
	}
}
