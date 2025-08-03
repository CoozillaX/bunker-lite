package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// AllowListDeleteRequest ..
type AllowListDeleteRequest struct {
	Token              string `json:"token,omitempty"`
	RentalServerNumber string `json:"rental_server_number,omitempty"`
	EulogistUserName   string `json:"eulogist_user_name,omitempty"`
}

// AllowListDeleteResponse ..
type AllowListDeleteResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// DeleteAllowListConfig ..
func DeleteAllowListConfig(c *gin.Context) {
	var request AllowListDeleteRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: fmt.Sprintf("DeleteAllowListConfig: 将目标用户从列表中删除时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: "DeleteAllowListConfig: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByName(request.Token, true)

	if len(request.RentalServerNumber) == 0 {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: "DeleteAllowListConfig: 提供的租赁服号不得为空",
			Success:   false,
		})
		return
	}

	if len(request.EulogistUserName) == 0 {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: "DeleteAllowListConfig: 提供的赞颂者用户名不得为空",
			Success:   false,
		})
		return
	}

	if !slices.Contains(user.RentalServerCanManage, request.RentalServerNumber) {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: fmt.Sprintf(
				"DeleteAllowListConfig: 目标租赁服 (%s) 不是您可以管理的租赁服, 看看是不是被管理员移除管理权限了？",
				request.RentalServerNumber,
			),
			Success: false,
		})
		return
	}

	if !database.CheckUserByName(request.EulogistUserName, true) {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: fmt.Sprintf(
				"DeleteAllowListConfig: 要操作的赞颂者用户 (%s) 没有找到, 看看他是不是刚刚改名了？",
				request.EulogistUserName,
			),
			Success: false,
		})
		return
	}
	dst := database.GetUserByName(request.EulogistUserName, true)

	findDstUser := false
	newConfig := make([]define.AllowListConfig, 0)
	configs := database.GetAllowServerConfig(request.RentalServerNumber, true)
	for _, value := range configs {
		if value.EulogistUserUniqueID == dst.UserUniqueID {
			findDstUser = true
			continue
		}
		newConfig = append(newConfig, value)
	}

	if !findDstUser {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: fmt.Sprintf(
				"DeleteAllowListConfig: 目标用户 (%s) 已经不在您租赁服 (%s) 的授权列表中",
				request.EulogistUserName,
				request.RentalServerNumber,
			),
			Success: false,
		})
		return
	}

	err = database.SetAllowServerConfig(request.RentalServerNumber, newConfig, true)
	if err != nil {
		c.JSON(http.StatusOK, AllowListDeleteResponse{
			ErrorInfo: fmt.Sprintf("DeleteAllowListConfig: 将目标用户从列表中删除时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, AllowListDeleteResponse{Success: true})
}
