package eulogist_api

import (
	"bunker-lite/database"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// AllowListGetRequest ..
type AllowListGetRequest struct {
	Token              string `json:"token,omitempty"`
	RentalServerNumber string `json:"rental_server_number,omitempty"`
}

// AllowListGetResponse ..
type AllowListGetResponse struct {
	ErrorInfo                string   `json:"error_info"`
	Success                  bool     `json:"success"`
	UserNames                []string `json:"user_names"`
	DisableOpertorVerify     []bool   `json:"disbale_operator_verify"`
	CanGetGameSavesKeyCipher []bool   `json:"can_get_game_saves_key_cipher"`
}

// GetAllowListConfig ..
func GetAllowListConfig(c *gin.Context) {
	var request AllowListGetRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, AllowListGetResponse{
			ErrorInfo: fmt.Sprintf("GetAllowListConfig: 获取权限列表时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, AllowListGetResponse{
			ErrorInfo: "GetAllowListConfig: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByName(request.Token, true)

	if len(request.RentalServerNumber) == 0 {
		c.JSON(http.StatusOK, AllowListGetResponse{
			ErrorInfo: "GetAllowListConfig: 提供的租赁服号不得为空",
			Success:   false,
		})
		return
	}

	if !slices.Contains(user.RentalServerCanManage, request.RentalServerNumber) {
		c.JSON(http.StatusOK, AllowListGetResponse{
			ErrorInfo: fmt.Sprintf(
				"GetAllowListConfig: 目标租赁服 (%s) 不是您可以管理的租赁服, 看看是不是被管理员移除管理权限了？",
				request.RentalServerNumber,
			),
			Success: false,
		})
		return
	}

	serverResp := AllowListGetResponse{Success: true}
	configs := database.GetAllowServerConfig(request.RentalServerNumber, true)
	for _, value := range configs {
		user := database.GetUserByUniqueID(value.EulogistUserUniqueID, true)
		serverResp.UserNames = append(serverResp.UserNames, user.UserName)
		serverResp.DisableOpertorVerify = append(serverResp.DisableOpertorVerify, user.DisableGlobalOpertorVerify)
		serverResp.CanGetGameSavesKeyCipher = append(serverResp.CanGetGameSavesKeyCipher, user.CanGetGameSavesKeyCipher)
	}

	c.JSON(http.StatusOK, serverResp)
}
