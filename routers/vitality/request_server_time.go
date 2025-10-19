package vitality_api

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-lite/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServerTimeRequest ..
type ServerTimeRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// ServerTimeResponse ..
type ServerTimeResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// RequestServerTime ..
func RequestServerTime(c *gin.Context) {
	var request ServerTimeRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, ServerTimeResponse{
			ErrorInfo: fmt.Sprintf("RequestServerTime: %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, ServerTimeResponse{
			ErrorInfo: "RequestServerTime: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	gu, activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, ServerTimeResponse{
			ErrorInfo: fmt.Sprintf("RequestServerTime: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, ServerTimeResponse{
			ErrorInfo: "RequestServerTime: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if request.SessionID != activeGu.SessionID {
		c.JSON(http.StatusOK, ServerTimeResponse{
			ErrorInfo: fmt.Sprintf(
				"RequestServerTime: Session ID not matched (expect = %#v, provided = %#v)",
				activeGu.SessionID, request.SessionID,
			),
			Success: false,
		})
		return
	}

	_, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.WebServerUrl + "/server-time").
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		_ = database.DeleteActiveG79User(request.Token, false)
		c.JSON(http.StatusOK, ServerTimeResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("RequestServerTime: %v", protocolError.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, ServerTimeResponse{Success: true})
}
