package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserInfoRequest ..
type UserInfoRequest struct {
	Token string `json:"token,omitempty"`
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
	user.UserPasswordSum256 = nil
	user.EulogistToken = ""

	c.JSON(http.StatusOK, UserInfoResponse{
		Success: true,
		Payload: define.EncodeEulogistUser(user),
	})
}
