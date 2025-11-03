package vitality_api

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-lite/service/database"
	"bunker-lite/service/routers/keys"
	"bunker-lite/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	RequestTypeGetCurrencyOnline uint8 = iota
	RequestTypeGetDailyGrowth
)

// VitalityDebugRequest ..
type VitalityDebugRequest struct {
	Token       string `json:"token,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	RequestType uint8  `json:"request_type,omitempty"`
}

// VitalityDebugResponse ..
type VitalityDebugResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`

	RestCurrencyTime int    `json:"rest_currency_time"`
	FormatDateString string `json:"format_date_string"`

	XpFromOnline   int `json:"xp_from_online"`
	XpFromRecharge int `json:"xp_from_recharge"`
}

// RequestVitalityDebug ..
func RequestVitalityDebug(c *gin.Context) {
	var request VitalityDebugRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, VitalityDebugResponse{
			ErrorInfo: fmt.Sprintf("RequestVitalityDebug: %v", err),
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
		c.JSON(http.StatusOK, VitalityDebugResponse{
			ErrorInfo: "RequestVitalityDebug: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.ActiveSessionTran.Lock(request.SessionID)
	defer database.ActiveSessionTran.Unlock(request.SessionID)

	session, found, err := database.LoadActiveSession(request.SessionID, true, true, false)
	if err != nil {
		c.JSON(http.StatusOK, VitalityDebugResponse{
			ErrorInfo: fmt.Sprintf("RequestVitalityDebug: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, VitalityDebugResponse{
			ErrorInfo: "RequestVitalityDebug: Session not found or is expired",
			Success:   false,
		})
		return
	}

	switch request.RequestType {
	case RequestTypeGetCurrencyOnline:
		// send http request
		reader, protocolError := session.G79User().CreateHttpClient().
			SetMethod(http.MethodPost).
			SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/get-currency-online").
			SetRawBody([]byte("{}")).
			SetTokenMode(g79.TOKEN_MODE_NORMAL).
			Do()
		if protocolError != nil {
			c.JSON(http.StatusOK, VitalityDebugResponse{
				ErrorInfo: fmt.Sprintf("RequestVitalityDebug: %v", protocolError.Error()),
				Success:   false,
			})
			return
		}
		// parse netease server response
		var query struct {
			Entity struct {
				RestCurrencyTime int    `json:"rest_currency_time"`
				Date             string `json:"date"`
			} `json:"entity"`
		}
		if err = json.NewDecoder(reader).Decode(&query); err != nil {
			c.JSON(http.StatusOK, VitalityDebugResponse{
				ErrorInfo: fmt.Sprintf("RequestVitalityDebug: %v", protocolError.Error()),
				Success:   false,
			})
			return
		}
		// response user request
		c.JSON(http.StatusOK, VitalityDebugResponse{
			Success:          true,
			RestCurrencyTime: query.Entity.RestCurrencyTime,
			FormatDateString: query.Entity.Date,
		})
	case RequestTypeGetDailyGrowth:
		// send http request
		reader, protocolError := session.G79User().CreateHttpClient().
			SetMethod(http.MethodPost).
			SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/pe-get-daily-growth-info").
			SetTokenMode(g79.TOKEN_MODE_NORMAL).
			Do()
		if protocolError != nil {
			c.JSON(http.StatusOK, VitalityDebugResponse{
				ErrorInfo: fmt.Sprintf("RequestVitalityDebug: %v", protocolError.Error()),
				Success:   false,
			})
			return
		}
		// parse netease server response
		var query struct {
			Entity struct {
				XpFromOnline   int `json:"1"`
				XpFromRecharge int `json:"2"`
			} `json:"entity"`
		}
		if err = json.NewDecoder(reader).Decode(&query); err != nil {
			c.JSON(http.StatusOK, VitalityDebugResponse{
				ErrorInfo: fmt.Sprintf("RequestVitalityDebug: %v", protocolError.Error()),
				Success:   false,
			})
			return
		}
		// response user request
		c.JSON(http.StatusOK, VitalityDebugResponse{
			Success:        true,
			XpFromOnline:   query.Entity.XpFromOnline,
			XpFromRecharge: query.Entity.XpFromRecharge,
		})
	}

}
