package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// AllowListUpdateRequest ..
type AllowListUpdateRequest struct {
	Token                    string `json:"token,omitempty"`
	RentalServerNumber       string `json:"rental_server_number,omitempty"`
	EulogistUserName         string `json:"eulogist_user_name,omitempty"`
	DisableOpertorVerify     bool   `json:"disable_operator_verify"`
	CanGetGameSavesKeyCipher bool   `json:"can_get_game_saves_key_cipher"`
}

// UserSearchResponse ..
type AllowListUpdateResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// UpdateAllowListConfig ..
func UpdateAllowListConfig(c *gin.Context) {
	var request AllowListUpdateRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: fmt.Sprintf("UpdateAllowListConfig: 设置权限时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: "UpdateAllowListConfig: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if len(request.RentalServerNumber) == 0 {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: "UpdateAllowListConfig: 提供的租赁服号不得为空",
			Success:   false,
		})
		return
	}

	if len(request.EulogistUserName) == 0 {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: "UpdateAllowListConfig: 提供的赞颂者用户名不得为空",
			Success:   false,
		})
		return
	}

	if !slices.Contains(user.RentalServerCanManage, request.RentalServerNumber) {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: fmt.Sprintf(
				"UpdateAllowListConfig: 目标租赁服 (%s) 不是您可以管理的租赁服, 看看是不是被管理员移除管理权限了？",
				request.RentalServerNumber,
			),
			Success: false,
		})
		return
	}

	if !database.CheckUserByName(request.EulogistUserName, true) {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: fmt.Sprintf(
				"UpdateAllowListConfig: 要操作的赞颂者用户 (%s) 没有找到, 看看他是不是刚刚改名了？",
				request.EulogistUserName,
			),
			Success: false,
		})
		return
	}
	dst := database.GetUserByName(request.EulogistUserName, true)

	findDstUser := false
	configs := database.GetAllowServerConfig(request.RentalServerNumber, true)
	for index, value := range configs {
		if value.EulogistUserUniqueID == dst.UserUniqueID {
			configs[index].DisableOpertorVerify = request.DisableOpertorVerify
			configs[index].CanGetGameSavesKeyCipher = request.CanGetGameSavesKeyCipher
			findDstUser = true
			break
		}
	}
	if !findDstUser {
		configs = append(configs, define.AllowListConfig{
			EulogistUserUniqueID:     dst.UserUniqueID,
			DisableOpertorVerify:     request.DisableOpertorVerify,
			CanGetGameSavesKeyCipher: request.CanGetGameSavesKeyCipher,
		})
	}

	err = database.SetAllowServerConfig(request.RentalServerNumber, configs, true)
	if err != nil {
		c.JSON(http.StatusOK, AllowListUpdateResponse{
			ErrorInfo: fmt.Sprintf("UpdateAllowListConfig: 设置权限时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, AllowListUpdateResponse{Success: true})
}
