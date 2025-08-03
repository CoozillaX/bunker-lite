package routers

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var mu = new(sync.Mutex)

// handlerWithMutex ..
func handlerWithMutex(handler func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()
		handler(c)
	}
}

// InitRouter 初始化赞颂者和标准服务的 API
func InitRouter() *gin.Engine {
	router := gin.Default()
	router = initStdRouter(router)
	router = initEulogistRouter(router)
	return router
}
