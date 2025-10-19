package vitality_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-lite/database"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// KeepGuAliveRequest ..
type KeepGuAliveRequest struct {
	Token          string `json:"token,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	ForceOperation bool   `json:"force_operation,omitempty"`
}

// KeepGuAliveResponse ..
type KeepGuAliveResponse struct {
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
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorInfo: "KeepGuAlive: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	gu, activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorInfo: "KeepGuAlive: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if !request.ForceOperation && request.SessionID != activeGu.SessionID {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorInfo: fmt.Sprintf(
				"KeepGuAlive: Session ID not matched (expect = %#v, provided = %#v)",
				activeGu.SessionID, request.SessionID,
			),
			Success: false,
		})
		return
	}

	reader, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.CoreServerUrl + "/authentication/update").
		SetEncryptSuffix(0xc).
		Do()
	if protocolError != nil {
		_ = database.DeleteActiveG79User(request.Token, false)
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", protocolError.Error()),
		})
		return
	}

	var query struct {
		Entity struct {
			Token string `json:"token"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(reader.Bytes(), &query); err != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", protocolError.Error()),
		})
		return
	}

	_, activeGu, err = database.ExtendG79UserLifeTime(request.Token, query.Entity.Token, activeGu.SessionID, false)
	if err != nil {
		c.JSON(http.StatusOK, KeepGuAliveResponse{
			ErrorInfo: fmt.Sprintf("KeepGuAlive: %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, KeepGuAliveResponse{
		Success:           true,
		SessionExpireTime: activeGu.G79UserExpireTime,
	})
}
