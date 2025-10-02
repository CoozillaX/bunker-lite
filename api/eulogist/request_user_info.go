package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	RequestTypeGetUserInfoNormal uint8 = iota
	RequestTypeGetUserInfoAdmin
)

// UserInfoRequest ..
type UserInfoRequest struct {
	Token            string `json:"token,omitempty"`
	RequestType      uint8  `json:"request_type"`
	EulogistUserName string `json:"eulogist_user_name,omitempty"`
}

// UserInfoResponse ..
type UserInfoResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
	Payload   []byte `json:"payload"`
}

// RequestUserInfo ..
func RequestUserInfo(c *gin.Context) {
	var request UserInfoRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, UserInfoResponse{
			ErrorInfo: fmt.Sprintf("RequestUserInfo: 请求赞颂者账号信息时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, UserInfoResponse{
			ErrorInfo: "RequestUserInfo: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if request.RequestType == RequestTypeGetUserInfoNormal {
		user.UserPasswordSum256 = nil
		user.EulogistToken = ""
		c.JSON(http.StatusOK, UserInfoResponse{
			Success: true,
			Payload: define.EncodeEulogistUser(user),
		})
		return
	}

	switch user.UserPermissionLevel {
	case define.UserPermissionSystem:
	case define.UserPermissionAdmin:
	default:
		c.JSON(http.StatusOK, UserInfoResponse{
			ErrorInfo: "RequestUserInfo: 权限不足",
			Success:   false,
		})
		return
	}

	if len(request.EulogistUserName) == 0 {
		c.JSON(http.StatusOK, UserInfoResponse{
			ErrorInfo: "RequestUserInfo: 提供的赞颂者用户名不得为空",
			Success:   false,
		})
		return
	}

	if !database.CheckUserByName(request.EulogistUserName, true) {
		c.JSON(http.StatusOK, UserInfoResponse{
			ErrorInfo: fmt.Sprintf(
				"RequestUserInfo: 目标用户 (%s) 没有找到, 他可能恰好改名了？",
				request.EulogistUserName,
			),
			Success: false,
		})
		return
	}

	dst := database.GetUserByName(request.EulogistUserName, true)
	dst.UserPasswordSum256 = nil
	dst.EulogistToken = ""

	c.JSON(http.StatusOK, UserInfoResponse{
		Success: true,
		Payload: define.EncodeEulogistUser(dst),
	})
}
