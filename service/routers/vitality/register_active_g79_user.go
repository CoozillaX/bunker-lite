package vitality_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-lite/service/database"
	"bunker-lite/service/define"
	"bunker-lite/service/utils"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterActiveGuRequest ..
type RegisterActiveGuRequest struct {
	Token              string `json:"token,omitempty"`
	OverrideSession    bool   `json:"override_session,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data,omitempty"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data,omitempty"`
}

// RegisterActiveGuResponse ..
type RegisterActiveGuResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionID         string `json:"session_id"`
	SessionType       uint8  `json:"session_type"`
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

	if tokenBytes, err := hex.DecodeString(request.Token); err == nil {
		decrypted, err := utils.DecryptPKCS1v15(define.TokenEncryptKey, tokenBytes)
		if err == nil {
			request.Token = string(decrypted)
		}
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

	if request.OverrideSession {
		activeGu, err = database.RegisterActiveG79User(
			helper,
			gameinfo.DefaultEngineVersion,
			request.ProvidedPeAuthData,
			request.ProvidedSaAuthData,
			false,
		)
	} else {
		activeGu, err = database.LoadOrRegisterActiveG79User(
			helper,
			gameinfo.DefaultEngineVersion,
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
		SessionID:         activeGu.SessionID,
		SessionType:       activeGu.SessionType,
		SessionExpireTime: activeGu.SessionExpireTime,
	})
}
