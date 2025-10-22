package vitality_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ResponseTypeFindSession uint8 = iota
	ResponseTypeNoSession
)

// SessionInfoRequest ..
type SessionInfoRequest struct {
	Token string `json:"token,omitempty"`
}

// SessionInfoResponse ..
type SessionInfoResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	ResponseType      uint8  `json:"response_type"`
	SessionID         string `json:"session_id"`
	SessionType       uint8  `json:"session_type"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// RequestSessionInfo ..
func RequestSessionInfo(c *gin.Context) {
	var request SessionInfoRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, SessionInfoResponse{
			ErrorInfo: fmt.Sprintf("RequestSessionInfo: %v", err),
			Success:   false,
		})
		return
	}

	decrypted, err := utils.DecryptPKCS1v15(define.TokenEncryptKey, []byte(request.Token))
	if err == nil {
		request.Token = string(decrypted)
	}
	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, SessionInfoResponse{
			ErrorInfo: "RequestSessionInfo: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, SessionInfoResponse{
			ErrorInfo: fmt.Sprintf("RequestSessionInfo: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, SessionInfoResponse{
			Success:      true,
			ResponseType: ResponseTypeNoSession,
		})
		return
	}

	c.JSON(http.StatusOK, SessionInfoResponse{
		Success:           true,
		ResponseType:      ResponseTypeFindSession,
		SessionID:         activeGu.SessionID,
		SessionType:       activeGu.SessionType,
		SessionExpireTime: activeGu.SessionExpireTime,
	})
}
