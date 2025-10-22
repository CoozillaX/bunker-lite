package vitality_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/utils"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const SessionLifeLimitSecond = 60 * 60 * 12

const (
	KeepGuAliveErrorMeetError uint8 = iota
	KeepGuAliveErrorLifeLimit
)

// KeepGuAliveRequest ..
type KeepGuAliveRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// KeepGuAliveResponse ..
type KeepGuAliveResponse struct {
	ErrorType         uint8  `json:"error_type"`
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// KeepGuAlive ..
func KeepGuAlive(c *gin.Context) {
	var request KeepGuAliveRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", err),
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
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: "KeepGuAlive: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: "KeepGuAlive: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if request.SessionID != activeGu.SessionID {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf(
				"KeepGuAlive: Session ID not matched (expect = %#v, provided = %#v)",
				activeGu.SessionID, request.SessionID,
			),
			Success: false,
		})
		return
	}
	if time.Now().Unix()-activeGu.SessionStartTime >= SessionLifeLimitSecond {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorLifeLimit,
			Success:   false,
		})
		return
	}

	protocolError := activeGu.RecordG79UserData.Update()
	if protocolError != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", protocolError.Error()),
			Success:   false,
		})
		return
	}

	activeGu, err = database.ExtendG79UserLifeTime(request.Token, activeGu, false)
	if err != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorType: KeepGuAliveErrorMeetError,
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, KeepGuAliveResponse{
		Success:           true,
		SessionExpireTime: activeGu.SessionExpireTime,
	})
}
