package eulogist_api

import (
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	AccountTypeMpayUser uint8 = iota
	AccountTypePeAuth
	AccountTypeSaAuth
)

// HelperInfoRequest ..
type HelperInfoRequest struct {
	Token string `json:"token,omitempty"`
}

// HelperInfoResponse ..
type HelperInfoResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	AccountType       uint8  `json:"account_type"`
	AccountExpireTime int64  `json:"account_expire_time"`
	GameNickName      string `json:"game_nick_name"`
	G79UserUID        string `json:"g79_user_uid"`
}

// GetStdHelperInfo ..
func GetStdHelperInfo(c *gin.Context) {
	var request HelperInfoRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, HelperInfoResponse{
			ErrorInfo: fmt.Sprintf("GetStdHelperInfo: 请求 MC 账号信息时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, HelperInfoResponse{
			ErrorInfo: "GetStdHelperInfo: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	account, ok := user.CurrentAuthServerAccount.Value()
	if !ok {
		c.JSON(http.StatusOK, HelperInfoResponse{
			ErrorInfo: "GetStdHelperInfo: 当前没有正在使用的 MC 账号",
			Success:   false,
		})
		return
	}
	if !account.IsStdAccount() {
		c.JSON(http.StatusOK, HelperInfoResponse{
			ErrorInfo: "GetStdHelperInfo: 当前正在使用的 MC 账号未被存放在标准验证服务中",
			Success:   false,
		})
		return
	}

	activeGu, protocolError := database.GetHelperBasicInfo(account.AuthServerSecret(), true)
	if protocolError != nil {
		c.JSON(http.StatusOK, HelperInfoResponse{
			ErrorInfo: fmt.Sprintf(
				"GetStdHelperInfo: 请求 MC 账号信息时出现问题, 原因是 %s",
				protocolError.Error(),
			),
			Success: false,
		})
		return
	}

	account.UpdateData(map[string]any{
		"gameNickName":       activeGu.RecordG79UserData.Username,
		"g79UserUID":         activeGu.RecordG79UserData.EntityID,
		"authHelperUniqueID": account.AuthServerSecret(),
	})
	user.CurrentAuthServerAccount = protocol.Option(account)

	for index, value := range user.MultipleAuthServerAccounts {
		if !value.IsStdAccount() {
			continue
		}
		if value.AuthServerSecret() == account.AuthServerSecret() {
			user.MultipleAuthServerAccounts[index] = account
		}
	}

	err = database.UpdateUserInfo(user, true)
	if !account.IsStdAccount() {
		c.JSON(http.StatusOK, HelperInfoResponse{
			ErrorInfo: fmt.Sprintf("GetStdHelperInfo: 请求 MC 账号信息时出现问题, 原因是 %s", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, HelperInfoResponse{
		Success:           true,
		AccountType:       activeGu.SessionType,
		AccountExpireTime: activeGu.SessionExpireTime,
		GameNickName:      activeGu.RecordG79UserData.Username,
		G79UserUID:        activeGu.RecordG79UserData.EntityID,
	})
}
