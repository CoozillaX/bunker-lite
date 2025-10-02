package eulogist_api

import (
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DevTokenAsk ..
type DevTokenAsk struct {
	Token          string `json:"token,omitempty"`
	HelperUniqueID string `json:"helper_unique_id,omitempty"`
}

// DevTokenResp ..
type DevTokenResp struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
	Token     string `json:"token"`
}

// DeveloperAskToken ..
func DeveloperAskToken(c *gin.Context) {
	var request DevTokenAsk

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, DevTokenResp{
			ErrorInfo: fmt.Sprintf("DeveloperAskToken: 请求 STD-AUTH-TOKEN 时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, DevTokenResp{
			ErrorInfo: "DeveloperAskToken: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if !user.CanGetHelperToken {
		c.JSON(http.StatusOK, DevTokenResp{
			ErrorInfo: "DeveloperAskToken: 您没有下载 STD-AUTH-TOKEN 的权限, 请联系赞颂者管理员",
			Success:   false,
		})
		return
	}

	if len(request.HelperUniqueID) == 0 {
		c.JSON(http.StatusOK, DevTokenResp{
			ErrorInfo: "DeveloperAskToken: 辅助用户的唯一 ID 不得为空",
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByUniqueID(request.HelperUniqueID, true) {
		c.JSON(http.StatusOK, DevTokenResp{
			ErrorInfo: "DeveloperAskToken: 未能找到目标辅助用户",
			Success:   false,
		})
		return
	}
	helper := database.GetAuthHelperByUniqueID(request.HelperUniqueID, true)

	c.JSON(http.StatusOK, DevTokenResp{
		Success: true,
		Token:   helper.HelperToken,
	})
}
