package eulogist_api

import (
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserSearchRequest ..
type UserSearchRequest struct {
	Token        string `json:"token,omitempty"`
	FilterString string `json:"filter_string,omitempty"`
}

// UserSearchResponse ..
type UserSearchResponse struct {
	ErrorInfo   string   `json:"error_info"`
	Success     bool     `json:"success"`
	HitUserName []string `json:"hit_user_name"`
}

// SearchEulogistUser ..
func SearchEulogistUser(c *gin.Context) {
	var request UserSearchRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, UserSearchResponse{
			ErrorInfo: fmt.Sprintf("SearchEulogistUser: 搜索赞颂者账户时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, UserSearchResponse{
			ErrorInfo: "SearchEulogistUser: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}

	if len(request.FilterString) == 0 {
		c.JSON(http.StatusOK, UserSearchResponse{
			ErrorInfo: "SearchEulogistUser: 要搜索的用户名的长度不得为 0",
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, UserSearchResponse{
		Success:     true,
		HitUserName: database.ListUsers(request.FilterString, true),
	})
}
