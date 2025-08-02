package routers

import (
	"bunker-lite/eulogist_api"
	"net/http"

	"github.com/gin-gonic/gin"
)

// initEulogistRouter 初始化赞颂者服务的 API
func initEulogistRouter(router *gin.Engine) *gin.Engine {
	eulogistApiGroup := router.Group("/eulogist_api")
	{
		eulogistApiGroup.POST("/register_or_login", eulogist_api.RegisterOrLogin)
		eulogistApiGroup.POST("/request_user_info", eulogist_api.RequestUserInfo)
	}

	// No router
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}
