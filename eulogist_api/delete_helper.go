package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelperAddRequest ..
type HelperDeleteRequest struct {
	Token string `json:"token,omitempty"`
	Index uint   `json:"index"`
}

// HelperAddResponse ..
type HelperDeleteResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// DeleteHelper ..
func DeleteHelper(c *gin.Context) {
	var request HelperDeleteRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, HelperDeleteResponse{
			ErrorInfo: fmt.Sprintf("DeleteHelper: 删除已有的 MC 账号时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, HelperDeleteResponse{
			ErrorInfo: "DeleteHelper: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if int(request.Index) >= len(user.MultipleAuthServerAccounts) {
		c.JSON(http.StatusOK, HelperDeleteResponse{
			ErrorInfo: fmt.Sprintf(
				"DeleteHelper: 超出索引 (最大为 %d 但提供了 %d)",
				len(user.MultipleAuthServerAccounts)-1,
				request.Index,
			),
			Success: false,
		})
		return
	}

	src, ok := user.CurrentAuthServerAccount.Value()
	if ok {
		var tryDeleteCurrentAccount bool

		dst := user.MultipleAuthServerAccounts[request.Index]
		if src.IsStdAccount() == dst.IsStdAccount() {
			if src.IsStdAccount() {
				if src.AuthServerSecret() == dst.AuthServerSecret() {
					tryDeleteCurrentAccount = true
				}
			} else {
				if src.AuthServerAddress() == dst.AuthServerAddress() && src.AuthServerSecret() == dst.AuthServerSecret() {
					tryDeleteCurrentAccount = true
				}
			}
		}

		if tryDeleteCurrentAccount {
			c.JSON(http.StatusOK, HelperDeleteResponse{
				ErrorInfo: "DeleteHelper: 当前账户正被您使用, 您必须切换当前使用的账户到其他账户后, 才能删除这个账户",
				Success:   false,
			})
			return
		}
	}

	newAccounts := make([]define.AuthServerAccount, 0)
	for index, value := range user.MultipleAuthServerAccounts {
		if index != int(request.Index) {
			newAccounts = append(newAccounts, value)
		}
	}
	user.MultipleAuthServerAccounts = newAccounts

	err = database.UpdateUserInfo(user, true)
	if err != nil {
		c.JSON(http.StatusOK, HelperDeleteResponse{
			ErrorInfo: fmt.Sprintf("DeleteHelper: 删除已有的 MC 账号时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, HelperDeleteResponse{Success: true})
}
