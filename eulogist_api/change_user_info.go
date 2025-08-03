package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"crypto/sha256"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserInfoChangeRequest ..
type UserInfoChangeRequest struct {
	Token             string `json:"token,omitempty"`
	NewName           string `json:"new_name,omitempty"`
	NewPasswordSum256 []byte `json:"new_password_sum256,omitempty"`
}

// UserInfoChangeResponse ..
type UserInfoChangeResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
	NewToken  string `json:"new_token"`
}

// ChangeUserInfo ..
func ChangeUserInfo(c *gin.Context) {
	var request UserInfoChangeRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, UserInfoChangeResponse{
			ErrorInfo: fmt.Sprintf("ChangeUserInfo: 更改赞颂者账户信息时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, UserInfoChangeResponse{
			ErrorInfo: "ChangeUserInfo: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}

	user := database.GetUserByToken(request.Token, true)
	user.UserName = request.NewName
	if len(request.NewPasswordSum256) > 0 {
		emptyUserPasswordSum256 := sha256.Sum256([]byte(define.UserPasswordSlat))
		if slices.Equal(request.NewPasswordSum256, emptyUserPasswordSum256[:]) {
			c.JSON(http.StatusOK, UserInfoChangeResponse{
				ErrorInfo: "ChangeUserInfo: 新密码不得为空",
				Success:   false,
			})
			return
		}
		user.UserPasswordSum256 = request.NewPasswordSum256
		user.EulogistToken = uuid.NewString()
	}

	err = database.UpdateUserInfo(user, true)
	if err != nil {
		c.JSON(http.StatusOK, UserInfoChangeResponse{
			ErrorInfo: fmt.Sprintf("ChangeUserInfo: 更改赞颂者账户信息时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, UserInfoChangeResponse{
		Success:  true,
		NewToken: user.EulogistToken,
	})
}
