package eulogist_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	RequestTypeRegisterSession uint8 = iota
	RequestTypeCleanUpSession
	RequestTypeEnableVitality
	RequestTypeDisableVitality
)

// LoginStatesSetRequest ..
type LoginStatesSetRequest struct {
	Token       string `json:"token,omitempty"`
	RequestType uint8  `json:"request_type,omitempty"`
	PeAuth      string `json:"pe_auth,omitempty"`
	SaAuth      string `json:"sa_auth,omitempty"`
}

// LoginStatesSetResponse ..
type LoginStatesSetResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// SetLoginStates ..
func SetLoginStates(c *gin.Context) {
	var request LoginStatesSetRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, LoginStatesSetResponse{
			ErrorInfo: fmt.Sprintf("SetLoginStates: 设置登录状态时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, LoginStatesSetResponse{
			ErrorInfo: "SetLoginStates: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	account, found := user.CurrentAuthServerAccount.Value()
	if !found {
		c.JSON(http.StatusOK, LoginStatesSetResponse{
			ErrorInfo: "SetLoginStates: 请确保您当前已经设置了正常使用的 MC 账号",
			Success:   false,
		})
		return
	}
	if !account.IsStdAccount() {
		c.JSON(http.StatusOK, LoginStatesSetResponse{
			ErrorInfo: "SetLoginStates: 请确保当前使用的 MC 账号来源于标准验证服务",
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByUniqueID(account.AuthServerSecret(), true) {
		c.JSON(http.StatusOK, LoginStatesSetResponse{
			ErrorInfo: fmt.Sprintf(
				"SetLoginStates: 不一致的底层数据库视图 (发生在 %v)",
				account.AuthServerSecret(),
			),
			Success: false,
		})
		return
	}
	helper := database.GetAuthHelperByUniqueID(account.AuthServerSecret(), true)

	switch request.RequestType {
	case RequestTypeRegisterSession:
		if len(request.PeAuth) == 0 && len(request.SaAuth) == 0 {
			c.JSON(http.StatusOK, LoginStatesSetResponse{
				ErrorInfo: "SetLoginStates: 提供的 Pe Auth 或 Sa Auth 的长度不得为 0",
				Success:   false,
			})
			return
		}
		_, err = database.RegisterActiveG79User(
			helper,
			gameinfo.DefaultEngineVersion,
			request.PeAuth,
			request.SaAuth,
			true,
		)
	case RequestTypeCleanUpSession:
		err = database.DeleteActiveG79User(helper.HelperToken, true)
	case RequestTypeEnableVitality:
		helper.EnableVitality = true
		err = database.UpdateHelperInfo(helper, true)
	case RequestTypeDisableVitality:
		helper.EnableVitality = false
		err = database.UpdateHelperInfo(helper, true)
	}
	if err != nil {
		c.JSON(http.StatusOK, LoginStatesSetResponse{
			ErrorInfo: fmt.Sprintf("SetLoginStates: 设置登录状态时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, LoginStatesSetResponse{Success: true})
}
