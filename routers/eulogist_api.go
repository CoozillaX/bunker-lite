package routers

import (
	"bunker-lite/eulogist_api"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var mu = new(sync.Mutex)

func handlerWithMutex(handler func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()
		handler(c)
	}
}

// initEulogistRouter 初始化赞颂者服务的 API
func initEulogistRouter(router *gin.Engine) *gin.Engine {
	eulogistApiGroup := router.Group("/eulogist_api")

	// Baisc
	{
		eulogistApiGroup.POST("/register_or_login", handlerWithMutex(eulogist_api.RegisterOrLogin))
		eulogistApiGroup.POST("/request_user_info", handlerWithMutex(eulogist_api.RequestUserInfo))
		eulogistApiGroup.POST("/search_eulogist_user", handlerWithMutex(eulogist_api.SearchEulogistUser))
		eulogistApiGroup.POST("/rental_server_list", handlerWithMutex(eulogist_api.RentalServerList))
	}

	// Helper
	{
		eulogistApiGroup.POST("/get_std_helper_info", handlerWithMutex(eulogist_api.GetStdHelperInfo))
		eulogistApiGroup.POST("/change_current_helper", handlerWithMutex(eulogist_api.ChangeCurrentHelper))
		eulogistApiGroup.POST("/add_helper_normal", handlerWithMutex(eulogist_api.AddHelperNormal))
		eulogistApiGroup.POST("/add_std_helper_sms", handlerWithMutex(eulogist_api.AddStdHelperSMS))
		eulogistApiGroup.POST("/modify_custom_helper", handlerWithMutex(eulogist_api.ModifyCustomHelper))
		eulogistApiGroup.POST("/delete_helper", handlerWithMutex(eulogist_api.DeleteHelper))
		eulogistApiGroup.POST("/dev_ask_token", handlerWithMutex(eulogist_api.DeveloperAskToken))
		eulogistApiGroup.POST("/set_pe_auth", handlerWithMutex(eulogist_api.SetPEAuth))
	}

	// Rental Server Manage
	{
		eulogistApiGroup.POST("/update_allow_list_config", handlerWithMutex(eulogist_api.UpdateAllowListConfig))
		eulogistApiGroup.POST("/get_allow_list_config", handlerWithMutex(eulogist_api.GetAllowListConfig))
		eulogistApiGroup.POST("/delete_allow_list_config", handlerWithMutex(eulogist_api.DeleteAllowListConfig))
	}

	// Eulogist Admin
	{
		eulogistApiGroup.POST("/admin_change_main_config", handlerWithMutex(eulogist_api.ChangeMainConfig))
		eulogistApiGroup.POST("/admin_change_manager", handlerWithMutex(eulogist_api.ChangeManager))
	}

	// No router
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}
