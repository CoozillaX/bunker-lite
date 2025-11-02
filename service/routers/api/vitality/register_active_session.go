package vitality_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-lite/service/database"
	"bunker-lite/service/define"
	"bunker-lite/service/routers/keys"
	"bunker-lite/utils"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterActiveSessionRequest ..
type RegisterActiveSessionRequest struct {
	Token              string `json:"token,omitempty"`
	OverrideSession    bool   `json:"override_session,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data,omitempty"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data,omitempty"`
}

// RegisterActiveSessionResponse ..
type RegisterActiveSessionResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionID         string `json:"session_id"`
	SessionType       uint8  `json:"session_type"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// RegisterActiveSession ..
func RegisterActiveSession(c *gin.Context) {
	var request RegisterActiveSessionRequest
	var session define.ActiveSession

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, RegisterActiveSessionResponse{
			ErrorInfo: fmt.Sprintf("RegisterActiveSession: %v", err),
			Success:   false,
		})
		return
	}

	if tokenBytes, err := hex.DecodeString(request.Token); err == nil {
		decrypted, err := utils.DecryptPKCS1v15(keys.TokenEncryptKey, tokenBytes)
		if err == nil {
			request.Token = string(decrypted)
		}
	}
	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, RegisterActiveSessionResponse{
			ErrorInfo: "RegisterActiveSession: Invalid token was provided",
			Success:   false,
		})
		return
	}
	helper := database.GetAuthHelperByToken(request.Token, true)

	if request.OverrideSession {
		session, err = database.RegisterActiveSession(
			helper,
			gameinfo.DefaultEngineVersion,
			request.ProvidedPeAuthData,
			request.ProvidedSaAuthData,
			true,
			true,
		)
	} else {
		session, err = database.LoadOrRegisterActiveSession(
			helper,
			gameinfo.DefaultEngineVersion,
			request.ProvidedPeAuthData,
			request.ProvidedSaAuthData,
			true,
			true,
			true,
		)
	}
	if err != nil {
		c.JSON(http.StatusOK, RegisterActiveSessionResponse{
			ErrorInfo: fmt.Sprintf("RegisterActiveSession: %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, RegisterActiveSessionResponse{
		Success:           true,
		SessionID:         session.SessionID,
		SessionType:       session.SessionType,
		SessionExpireTime: session.SessionExpireTime,
	})
}
