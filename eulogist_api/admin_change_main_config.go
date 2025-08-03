package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChangeMainConfigRequest ..
type ChangeMainConfigRequest struct {
	Token            string `json:"token,omitempty"`
	EulogistUserName string `json:"eulogist_user_name,omitempty"`
	NewUserData      []byte `json:"new_user_data"`
}

// ChangeMainConfigResponse ..
type ChangeMainConfigResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// ChangeMainConfig ..
func ChangeMainConfig(c *gin.Context) {
	var request ChangeMainConfigRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, ChangeMainConfigResponse{
			ErrorInfo: fmt.Sprintf("ChangeMainConfig: 更改用户主要配置时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, ChangeMainConfigResponse{
			ErrorInfo: "ChangeMainConfig: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	switch user.UserPermissionLevel {
	case define.UserPermissionSystem:
	case define.UserPermissionAdmin:
	default:
		c.JSON(http.StatusOK, ChangeMainConfigResponse{
			ErrorInfo: "ChangeMainConfig: 权限不足",
			Success:   false,
		})
		return
	}

	if len(request.EulogistUserName) == 0 {
		c.JSON(http.StatusOK, ChangeMainConfigResponse{
			ErrorInfo: "ChangeMainConfig: 提供的赞颂者用户名不得为空",
			Success:   false,
		})
		return
	}

	if !database.CheckUserByName(request.EulogistUserName, true) {
		c.JSON(http.StatusOK, ChangeMainConfigResponse{
			ErrorInfo: fmt.Sprintf(
				"ChangeMainConfig: 目标用户 (%s) 没有找到, 他可能恰好改名了？",
				request.EulogistUserName,
			),
			Success: false,
		})
		return
	}

	databaseUser := database.GetUserByName(request.EulogistUserName, true)
	providedUser := define.DecodeEulogistUser(request.NewUserData)

	if databaseUser.CanGetHelperToken && !providedUser.CanGetHelperToken {
		for _, value := range databaseUser.MultipleAuthServerAccounts {
			if !value.IsStdAccount() {
				continue
			}
			if err = database.UpdateHelperToken(value.AuthServerSecret(), true); err != nil {
				c.JSON(http.StatusOK, ChangeMainConfigResponse{
					ErrorInfo: fmt.Sprintf("ChangeMainConfig: 更改用户主要配置时出现问题, 原因是 %v", err),
					Success:   false,
				})
				return
			}
		}
	}

	databaseUser.UserName = providedUser.UserName
	databaseUser.UserPermissionLevel = providedUser.UserPermissionLevel
	databaseUser.UnbanUnixTime = providedUser.UnbanUnixTime
	databaseUser.DisableGlobalOpertorVerify = providedUser.DisableGlobalOpertorVerify
	databaseUser.CanAccessAnyRentalServer = providedUser.CanAccessAnyRentalServer
	databaseUser.CanGetGameSavesKeyCipher = providedUser.CanGetGameSavesKeyCipher
	databaseUser.CanGetHelperToken = providedUser.CanGetHelperToken

	err = database.UpdateUserInfo(databaseUser, true)
	if err != nil {
		c.JSON(http.StatusOK, ChangeMainConfigResponse{
			ErrorInfo: fmt.Sprintf("ChangeMainConfig: 更改用户主要配置时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, ChangeMainConfigResponse{Success: true})
}
