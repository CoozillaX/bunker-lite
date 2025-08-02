package routers

import (
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化赞颂者和标准服务的 API
func InitRouter() *gin.Engine {
	router := gin.Default()
	router = initEulogistRouter(router)
	return router
}
