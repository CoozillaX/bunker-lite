package eulogist_api

import (
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// HelperChangeRequest ..
type HelperChangeRequest struct {
	Token string `json:"token,omitempty"`
	Index uint   `json:"index"`
}

// HelperChangeResponse ..
type HelperChangeResponse struct {
	ErrorInfo            string `json:"error_info"`
	NetEaseRequireVerify bool   `json:"netease_require_verify"`
	VerifyURL            string `json:"verify_url"`
	Success              bool   `json:"success"`
	GameNickName         string `json:"game_nick_name"`
	G79UserUID           string `json:"g79_user_uid"`
}

// ChangeCurrentHelper ..
func ChangeCurrentHelper(c *gin.Context) {
	var request HelperChangeRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, HelperChangeResponse{
			ErrorInfo: fmt.Sprintf("ChangeCurrentHelper: 切换 MC 账号时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, HelperChangeResponse{
			ErrorInfo: "ChangeCurrentHelper: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if int(request.Index) >= len(user.MultipleAuthServerAccounts) {
		c.JSON(http.StatusOK, HelperChangeResponse{
			ErrorInfo: fmt.Sprintf(
				"ChangeCurrentHelper: 超出索引 (最大为 %d 但提供了 %d)",
				len(user.MultipleAuthServerAccounts)-1,
				request.Index,
			),
			Success: false,
		})
		return
	}

	account := user.MultipleAuthServerAccounts[request.Index]
	if !account.IsStdAccount() {
		user.CurrentAuthServerAccount = protocol.Option(account)
		if err = database.UpdateUserInfo(user, true); err != nil {
			c.JSON(http.StatusOK, HelperChangeResponse{
				ErrorInfo: fmt.Sprintf("ChangeCurrentHelper: 切换 MC 账号时出现问题, 原因是 %v", err),
				Success:   false,
			})
			return
		}
		c.JSON(http.StatusOK, HelperChangeResponse{
			Success: true,
		})
		return
	}

	nickName, g79UserUID, protocolError := database.GetHelperBasicInfo(account.AuthServerSecret(), true)
	if protocolError != nil {
		c.JSON(http.StatusOK, HelperChangeResponse{
			ErrorInfo:            fmt.Sprintf("ChangeCurrentHelper: 切换 MC 账号时出现问题, 原因是 %s", protocolError.Error()),
			NetEaseRequireVerify: len(protocolError.VerifyUrl) != 0,
			VerifyURL:            protocolError.VerifyUrl,
			Success:              false,
		})
		return
	}

	account.UpdateData(map[string]any{
		"gameNickName":       nickName,
		"g79UserUID":         g79UserUID,
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
	if err != nil {
		c.JSON(http.StatusOK, HelperChangeResponse{
			ErrorInfo: fmt.Sprintf("ChangeCurrentHelper: 切换 MC 账号时出现问题, 原因是 %s", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, HelperChangeResponse{
		Success:      true,
		GameNickName: nickName,
		G79UserUID:   g79UserUID,
	})
}
