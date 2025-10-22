package routers

import (
	eulogist_api "bunker-lite/routers/eulogist"
	std_api "bunker-lite/routers/standard"
	vitality_api "bunker-lite/routers/vitality"
	"net/http"

	"github.com/gin-gonic/gin"
)

// initStdRouter 初始化 Phoenix 标准接口的 API
func initStdRouter(router *gin.Engine) *gin.Engine {
	stdApiGroup := router.Group("/api")

	// Phoenix Auth API (mv4, bunker, v2.7)
	{
		stdApiGroup.GET("/new", handlerWithMutex(std_api.New))
		stdApiGroup.POST("/phoenix/login", handlerWithMutex(std_api.Login))
		stdApiGroup.POST("/phoenix/tan_lobby_login", handlerWithMutex(std_api.TanLobbyLogin))
		stdApiGroup.POST("/phoenix/tan_lobby_create", handlerWithMutex(std_api.TanLobbyCreate))
		stdApiGroup.POST("/phoenix/tan_lobby_debug", handlerWithMutex(std_api.TanLobbyDebug))
		stdApiGroup.POST("/phoenix/transfer_check_num", handlerWithMutex(std_api.TransferCheckNum))
		stdApiGroup.GET("/phoenix/transfer_start_type", handlerWithMutex(std_api.TransferStartType))
	}

	// No router
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}

// initEulogistRouter 初始化赞颂者服务的 API
func initEulogistRouter(router *gin.Engine) *gin.Engine {
	eulogistApiGroup := router.Group("/eulogist_api")

	// Basic
	{
		eulogistApiGroup.POST("/register_or_login", handlerWithMutex(eulogist_api.RegisterOrLogin))
		eulogistApiGroup.POST("/request_user_info", handlerWithMutex(eulogist_api.RequestUserInfo))
		eulogistApiGroup.POST("/search_eulogist_user", handlerWithMutex(eulogist_api.SearchEulogistUser))
		eulogistApiGroup.POST("/rental_server_list", handlerWithMutex(eulogist_api.RentalServerList))
		eulogistApiGroup.POST("/change_user_info", handlerWithMutex(eulogist_api.ChangeUserInfo))
		eulogistApiGroup.POST("/get_game_saves_key", handlerWithMutex(eulogist_api.GetGameSavesKey))
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
		eulogistApiGroup.POST("/set_login_states", handlerWithMutex(eulogist_api.SetLoginStates))
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

	// Develop features
	{
		eulogistApiGroup.POST("/get_built_in_skin", handlerWithMutex(eulogist_api.GetBuiltInSkin))
		eulogistApiGroup.POST("/set_skin_cache", handlerWithMutex(eulogist_api.SetSkinCache))
	}

	// No router
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}

// initVitalityRouter ..
func initVitalityRouter(router *gin.Engine) *gin.Engine {
	vitalityApiGroup := router.Group("/vitality_api")

	// Vitality API
	{
		vitalityApiGroup.POST("/registry_active_gu", handlerWithMutex(vitality_api.RegisterActiveGu))
		vitalityApiGroup.POST("/request_session_info", vitality_api.RequestSessionInfo)
		vitalityApiGroup.POST("/clean_up_session", vitality_api.CleanUpSession)
		vitalityApiGroup.POST("/request_daily_growth", vitality_api.RequestDailyGrowth)
		vitalityApiGroup.POST("/get_currency_online", vitality_api.GetCurrencyOnline)
		vitalityApiGroup.POST("/keep_gu_alive", vitality_api.KeepGuAlive)
	}

	// No router
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}
