package vitality_api

import (
	"bunker-lite/database"
	"bunker-lite/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CleanUpSessionRequest ..
type CleanUpSessionRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// CleanUpSessionResponse ..
type CleanUpSessionResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// CleanUpSession ..
func CleanUpSession(c *gin.Context) {
	var request CleanUpSessionRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, CleanUpSessionResponse{
			ErrorInfo: fmt.Sprintf("CleanUpSession: %v", err),
			Success:   false,
		})
		return
	}

	decrypted, err := utils.DecryptPKCS1v15(TokenEncryptKey, []byte(request.Token))
	if err == nil {
		request.Token = string(decrypted)
	}
	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, CleanUpSessionResponse{
			ErrorInfo: "CleanUpSession: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, CleanUpSessionResponse{
			ErrorInfo: fmt.Sprintf("CleanUpSession: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, CleanUpSessionResponse{
			ErrorInfo: "CleanUpSession: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if request.SessionID != activeGu.SessionID {
		c.JSON(http.StatusOK, CleanUpSessionResponse{
			ErrorInfo: fmt.Sprintf(
				"CleanUpSession: Session ID not matched (expect = %#v, provided = %#v)",
				activeGu.SessionID, request.SessionID,
			),
			Success: false,
		})
		return
	}

	err = database.DeleteActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, CleanUpSessionResponse{
			ErrorInfo: fmt.Sprintf("CleanUpSession: %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, CleanUpSessionResponse{Success: true})
}
