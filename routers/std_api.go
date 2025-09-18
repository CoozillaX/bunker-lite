package routers

import (
	"bunker-lite/std_api"
	"net/http"

	"github.com/gin-gonic/gin"
)

// initStdRouter 初始化 Phoenix 标准接口的 API
func initStdRouter(router *gin.Engine) *gin.Engine {
	stdApiGroup := router.Group("/api")

	// Phoenix Auth API (mv4, bunker, v2.4)
	{
		stdApiGroup.GET("/new", handlerWithMutex(std_api.New))
		stdApiGroup.POST("/phoenix/login", handlerWithMutex(std_api.Login))
		stdApiGroup.POST("/phoenix/tan_lobby_login", handlerWithMutex(std_api.TanLobbyLogin))
		stdApiGroup.POST("/phoenix/tan_lobby_create", handlerWithMutex(std_api.TanLobbyCreate))
		stdApiGroup.POST("/phoenix/transfer_check_num", handlerWithMutex(std_api.TransferCheckNum))
		stdApiGroup.GET("/phoenix/transfer_start_type", handlerWithMutex(std_api.TransferStartType))
	}

	// No router
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}
