package vitality_api

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-lite/service/database"
	"bunker-lite/service/keys"
	"bunker-lite/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DailyGrowthRequest ..
type DailyGrowthRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// DailyGrowthResponse ..
type DailyGrowthResponse struct {
	ErrorInfo      string `json:"error_info"`
	Success        bool   `json:"success"`
	XpFromOnline   int    `json:"xp_from_online"`
	XpFromRecharge int    `json:"xp_from_recharge"`
}

// RequestDailyGrowth ..
func RequestDailyGrowth(c *gin.Context) {
	var request DailyGrowthRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: fmt.Sprintf("RequestDailyGrowth: %v", err),
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
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: "RequestDailyGrowth: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: fmt.Sprintf("RequestDailyGrowth: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: "RequestDailyGrowth: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if request.SessionID != activeGu.SessionID {
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: fmt.Sprintf(
				"RequestDailyGrowth: Session ID not matched (expect = %#v, provided = %#v)",
				activeGu.SessionID, request.SessionID,
			),
			Success: false,
		})
		return
	}

	reader, protocolError := activeGu.RecordG79UserData.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/pe-get-daily-growth-info").
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: fmt.Sprintf("RequestDailyGrowth: %v", protocolError.Error()),
			Success:   false,
		})
		return
	}

	var query struct {
		Entity struct {
			XpFromOnline   int `json:"1"`
			XpFromRecharge int `json:"2"`
		} `json:"entity"`
	}
	if err = json.NewDecoder(reader).Decode(&query); err != nil {
		c.JSON(http.StatusOK, DailyGrowthResponse{
			ErrorInfo: fmt.Sprintf("RequestDailyGrowth: %v", protocolError.Error()),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, DailyGrowthResponse{
		Success:        true,
		XpFromOnline:   query.Entity.XpFromOnline,
		XpFromRecharge: query.Entity.XpFromRecharge,
	})
}
