package vitality_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterActiveGuRequest ..
type RegisterActiveGuRequest struct {
	Token              string `json:"token,omitempty"`
	OverrideSession    bool   `json:"override_session,omitempty"`
	EngineVersion      string `json:"engine_version,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data,omitempty"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data,omitempty"`
}

// RegisterActiveGuResponse ..
type RegisterActiveGuResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionType       uint8  `json:"session_type"`
	SessionID         string `json:"session_id"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// RegisterActiveGu ..
func RegisterActiveGu(c *gin.Context) {
	var request RegisterActiveGuRequest
	var activeGu define.ActiveG79User

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, RegisterActiveGuResponse{
			ErrorInfo: fmt.Sprintf("RegisterActiveGu: %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, RegisterActiveGuResponse{
			ErrorInfo: "RegisterActiveGu: Invalid token was provided",
			Success:   false,
		})
		return
	}

	helper := database.GetAuthHelperByToken(request.Token, true)
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	if len(request.EngineVersion) == 0 {
		request.EngineVersion = gameinfo.DefaultEngineVersion
	}
	if request.OverrideSession {
		_, activeGu, err = database.RegisterActiveG79User(
			helper,
			request.EngineVersion,
			request.ProvidedPeAuthData,
			request.ProvidedSaAuthData,
			false,
		)
	} else {
		_, activeGu, err = database.LoadOrRegisterActiveG79User(
			helper,
			request.EngineVersion,
			request.ProvidedPeAuthData,
			request.ProvidedSaAuthData,
			false,
		)
	}
	if err != nil {
		c.JSON(http.StatusOK, RegisterActiveGuResponse{
			ErrorInfo: fmt.Sprintf("RegisterActiveGu: %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, RegisterActiveGuResponse{
		Success:           true,
		SessionType:       activeGu.SessionType,
		SessionID:         activeGu.SessionID,
		SessionExpireTime: activeGu.SessionExpireTime,
	})
}
