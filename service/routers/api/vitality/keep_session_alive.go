package vitality_api

import (
	"bunker-lite/service/database"
	"bunker-lite/service/define"
	"bunker-lite/service/routers/keys"
	"bunker-lite/utils"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	KeepSessionAliveErrorMeetError uint8 = iota
	KeepSessionAliveErrorLifeLimit
)

// KeepSessionAliveRequest ..
type KeepSessionAliveRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// KeepSessionAliveResponse ..
type KeepSessionAliveResponse struct {
	ErrorType         uint8  `json:"error_type"`
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// KeepSessionAlive ..
func KeepSessionAlive(c *gin.Context) {
	var request KeepSessionAliveRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, KeepSessionAliveResponse{
			ErrorType: KeepSessionAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf("KeepSessionAlive: %v", err),
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
		c.JSON(http.StatusOK, KeepSessionAliveResponse{
			ErrorType: KeepSessionAliveErrorMeetError,
			ErrorInfo: "KeepSessionAlive: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.ActiveSessionTran.Lock(request.SessionID)
	defer database.ActiveSessionTran.Unlock(request.SessionID)

	session, found, err := database.LoadActiveSession(request.SessionID, true, true, false)
	if err != nil {
		c.JSON(http.StatusOK, KeepSessionAliveResponse{
			ErrorType: KeepSessionAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf("KeepSessionAlive: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, KeepSessionAliveResponse{
			ErrorType: KeepSessionAliveErrorMeetError,
			ErrorInfo: "KeepSessionAlive: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if time.Now().Unix()-session.SessionStartTime >= define.SessionMaxLifeTimeSecond {
		c.JSON(http.StatusOK, KeepSessionAliveResponse{
			ErrorType: KeepSessionAliveErrorLifeLimit,
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, KeepSessionAliveResponse{
		Success:           true,
		SessionExpireTime: session.SessionExpireTime,
	})
}
