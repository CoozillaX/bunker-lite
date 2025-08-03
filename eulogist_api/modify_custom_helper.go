package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CustomHelperModifyRequest ..
type CustomHelperModifyRequest struct {
	Token             string `json:"token,omitempty"`
	Index             uint   `json:"index"`
	AuthServerAddress string `json:"auth_server_address"`
	AuthServerToken   string `json:"auth_server_token"`
}

// CustomHelperModifyResponse ..
type CustomHelperModifyResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// ModifyCustomHelper ..
func ModifyCustomHelper(c *gin.Context) {
	var request CustomHelperModifyRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, CustomHelperModifyResponse{
			ErrorInfo: fmt.Sprintf("ModifyCustomHelper: 更新第三方验证账户时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, CustomHelperModifyResponse{
			ErrorInfo: "ModifyCustomHelper: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if int(request.Index) >= len(user.MultipleAuthServerAccounts) {
		c.JSON(http.StatusOK, CustomHelperModifyResponse{
			ErrorInfo: fmt.Sprintf(
				"ModifyCustomHelper: 超出索引 (最大为 %d 但提供了 %d)",
				len(user.MultipleAuthServerAccounts)-1,
				request.Index,
			),
			Success: false,
		})
		return
	}

	src := user.MultipleAuthServerAccounts[request.Index]
	if src.IsStdAccount() {
		c.JSON(http.StatusOK, CustomHelperModifyResponse{
			ErrorInfo: "ModifyCustomHelper: 内置验证服务的 MC 账户是不可修改的, 您只能删除后重新添加",
			Success:   false,
		})
		return
	}

	if len(request.AuthServerAddress) == 0 {
		c.JSON(http.StatusOK, CustomHelperModifyResponse{
			ErrorInfo: "ModifyCustomHelper: 验证服务地址的长度不得为 0",
			Success:   false,
		})
		return
	}

	newAccount := define.CustomAuthServerAccount{}
	newAccount.UpdateData(map[string]any{
		"internalAccountID": src.(*define.CustomAuthServerAccount).InternalAccountID(),
		"authServerAddress": request.AuthServerAddress,
		"authServerToken":   request.AuthServerToken,
	})
	user.MultipleAuthServerAccounts[request.Index] = &newAccount

	dst, ok := user.CurrentAuthServerAccount.Value()
	if ok {
		if !dst.IsStdAccount() && src.AuthServerAddress() == dst.AuthServerAddress() && src.AuthServerSecret() == dst.AuthServerSecret() {
			user.CurrentAuthServerAccount = protocol.Option(user.MultipleAuthServerAccounts[request.Index])
		}
	}

	err = database.UpdateUserInfo(user, true)
	if err != nil {
		c.JSON(http.StatusOK, CustomHelperModifyResponse{
			ErrorInfo: fmt.Sprintf("ModifyCustomHelper: 更新第三方验证账户时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, CustomHelperModifyResponse{Success: true})
}
