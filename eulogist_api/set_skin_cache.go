package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetSkinCacheRequest ..
type SetSkinCacheRequest struct {
	Token           string `json:"token,omitempty"`
	SkinDownloadURL string `json:"skin_download_url,omitempty"`
}

// SetSkinCacheResponse ..
type SetSkinCacheResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// SetSkinCache ..
func SetSkinCache(c *gin.Context) {
	var request SetSkinCacheRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, SetSkinCacheResponse{
			ErrorInfo: fmt.Sprintf("SetSkinCache: 设置皮肤缓存时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, SetSkinCacheResponse{
			ErrorInfo: "SetSkinCache: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}

	user := database.GetUserByToken(request.Token, true)
	switch user.UserPermissionLevel {
	case define.UserPermissionSystem:
	case define.UserPermissionAdmin:
	case define.UserPermissionAdvance:
	default:
		c.JSON(http.StatusOK, SetSkinCacheResponse{
			ErrorInfo: "SetSkinCache: 权限不足",
			Success:   false,
		})
		return
	}

	if len(request.SkinDownloadURL) == 0 {
		c.JSON(http.StatusOK, SetSkinCacheResponse{
			ErrorInfo: "SetSkinCache: SkinDownloadURL 不得为空",
			Success:   false,
		})
		return
	}

	if err = database.SetUserSkinCache(user.UserUniqueID, request.SkinDownloadURL, true); err != nil {
		c.JSON(http.StatusOK, SetSkinCacheResponse{
			ErrorInfo: fmt.Sprintf("SetSkinCache: 设置皮肤缓存时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, SetSkinCacheResponse{Success: true})
}
