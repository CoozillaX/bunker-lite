package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// LoginRequest ..
type LoginRequest struct {
	IsRegister         bool   `json:"is_register"`
	UserName           string `json:"user_name"`
	UserPasswordSum256 []byte `json:"user_password_sum256"`
}

// LoginResponse ..
type LoginResponse struct {
	ErrorInfo     string `json:"error_info"`
	Success       bool   `json:"success"`
	EulogistToken string `json:"eulogist_token"`
}

// RegisterOrLogin ..
func RegisterOrLogin(c *gin.Context) {
	var request LoginRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, LoginResponse{
			ErrorInfo: fmt.Sprintf("RegisterOrLogin: 注册或登录赞颂者账号时出现问题，原因是 %v", err),
			Success:   false,
		})
		return
	}

	if request.IsRegister {
		if len(request.UserName) > 32 {
			c.JSON(http.StatusOK, LoginResponse{
				ErrorInfo: "RegisterOrLogin: 用户名长度不得超过 32 个字符",
				Success:   false,
			})
			return
		}
		if strings.Contains(request.UserName, "\n") {
			c.JSON(http.StatusOK, LoginResponse{
				ErrorInfo: "RegisterOrLogin: 用户名不得包含换行符",
				Success:   false,
			})
			return
		}

		err = database.CreateUser(request.UserName, request.UserPasswordSum256, define.UserPermissionDefault, true)
		if err != nil {
			c.JSON(http.StatusOK, LoginResponse{
				ErrorInfo: fmt.Sprintf("RegisterOrLogin: 注册赞颂者账号时出现问题，原因是 %v", err),
				Success:   false,
			})
			return
		}

		user := database.GetUserByName(request.UserName, true)
		c.JSON(http.StatusOK, LoginResponse{
			Success:       true,
			EulogistToken: user.EulogistToken,
		})
		return
	}

	if !database.CheckUserByName(request.UserName, true) {
		c.JSON(http.StatusOK, LoginResponse{
			ErrorInfo: "RegisterOrLogin: 目标赞颂者用户不存在",
			Success:   false,
		})
		return
	}

	user := database.GetUserByName(request.UserName, true)
	if !slices.Equal(user.UserPasswordSum256, request.UserPasswordSum256) {
		c.JSON(http.StatusOK, LoginResponse{
			ErrorInfo: "RegisterOrLogin: 提供的用户名或密码不正确",
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success:       true,
		EulogistToken: user.EulogistToken,
	})
}
