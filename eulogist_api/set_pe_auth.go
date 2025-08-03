package eulogist_api

import (
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PEAuthSetRequest ..
type PEAuthSetRequest struct {
	Token  string `json:"token,omitempty"`
	PEAuth string `json:"pe_auth,omitempty"`
}

// PEAuthSetResponse ..
type PEAuthSetResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// DeleteHelper ..
func SetPEAuth(c *gin.Context) {
	var request PEAuthSetRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, PEAuthSetResponse{
			ErrorInfo: fmt.Sprintf("SetPEAuth: 设置 PE Auth 时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, PEAuthSetResponse{
			ErrorInfo: "SetPEAuth: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}

	if len(request.PEAuth) == 0 {
		c.JSON(http.StatusOK, PEAuthSetResponse{
			ErrorInfo: "SetPEAuth: 提供的 PE Auth 的长度不得为 0",
			Success:   false,
		})
		return
	}

	user := database.GetUserByToken(request.Token, true)
	user.ProvidedPeAuthData = request.PEAuth

	err = database.UpdateUserInfo(user, true)
	if err != nil {
		c.JSON(http.StatusOK, PEAuthSetResponse{
			ErrorInfo: fmt.Sprintf("SetPEAuth: 设置 PE Auth 时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, PEAuthSetResponse{Success: true})
}
