package eulogist_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthDataSetRequest ..
type AuthDataSetRequest struct {
	Token   string `json:"token,omitempty"`
	DoClean bool   `json:"do_clean,omitempty"`
	PEAuth  string `json:"pe_auth,omitempty"`
	SaAuth  string `json:"sa_auth,omitempty"`
}

// AuthDataSetResponse ..
type AuthDataSetResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// SetAuthData ..
func SetAuthData(c *gin.Context) {
	var request AuthDataSetRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: fmt.Sprintf("SetAuthData: 设置 PE Auth 时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: "SetAuthData: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	account, found := user.CurrentAuthServerAccount.Value()
	if !found {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: "SetAuthData: 当前没有设置 MC 账号 (标记 0)",
			Success:   false,
		})
		return
	}
	if !account.IsStdAccount() {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: "SetAuthData: 您必须确保当前使用的 MC 账号是 内置验证服务 的账号",
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByUniqueID(account.AuthServerSecret(), true) {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: "SetAuthData: 当前没有设置 MC 账号 (标记 1)",
			Success:   false,
		})
		return
	}
	helper := database.GetAuthHelperByUniqueID(account.AuthServerSecret(), true)

	if request.DoClean {
		if err = database.DeleteActiveG79User(helper.HelperToken, true); err != nil {
			c.JSON(http.StatusOK, AuthDataSetResponse{
				ErrorInfo: fmt.Sprintf("SetAuthData: 设置 PE Auth 时出现问题, 原因是 %v", err),
				Success:   false,
			})
			return
		}
		c.JSON(http.StatusOK, AuthDataSetResponse{Success: true})
		return
	}

	if len(request.PEAuth) == 0 || len(request.SaAuth) == 0 {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: "SetAuthData: 提供的 PE Auth 或 Sa Auth 的长度不得为 0",
			Success:   false,
		})
		return
	}
	_, _, err = database.RegisterActiveG79User(
		helper,
		gameinfo.DefaultEngineVersion,
		request.PEAuth,
		request.SaAuth,
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, AuthDataSetResponse{
			ErrorInfo: fmt.Sprintf("SetAuthData: 设置 PE Auth 时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, AuthDataSetResponse{Success: true})
}
