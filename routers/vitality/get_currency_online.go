package vitality_api

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-lite/database"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CurrencyOnlineRequest ..
type CurrencyOnlineRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// CurrencyOnlineResponse ..
type CurrencyOnlineResponse struct {
	ErrorInfo        string `json:"error_info"`
	Success          bool   `json:"success"`
	RestCurrencyTime int    `json:"rest_currency_time"`
	FormatDateString string `json:"format_date_string"`
}

// GetCurrencyOnline ..
func GetCurrencyOnline(c *gin.Context) {
	var request CurrencyOnlineRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			ErrorInfo: fmt.Sprintf("GetCurrencyOnline: %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.Token, true) {
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			ErrorInfo: "GetCurrencyOnline: Invalid token was provided",
			Success:   false,
		})
		return
	}
	database.LockG79Transaction(request.Token)
	defer database.UnlockG79Transaction(request.Token)

	gu, activeGu, found, err := database.LoadActiveG79User(request.Token, false)
	if err != nil {
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			ErrorInfo: fmt.Sprintf("GetCurrencyOnline: %v", err),
			Success:   false,
		})
		return
	}
	if !found {
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			ErrorInfo: "GetCurrencyOnline: Session not found or is expired",
			Success:   false,
		})
		return
	}
	if request.SessionID != activeGu.SessionID {
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			ErrorInfo: fmt.Sprintf(
				"GetCurrencyOnline: Session ID not matched (expect = %#v, provided = %#v)",
				activeGu.SessionID, request.SessionID,
			),
			Success: false,
		})
		return
	}

	reader, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/get-currency-online").
		SetRawBody([]byte("{}")).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		_ = database.DeleteActiveG79User(request.Token, false)
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("GetCurrencyOnline: %v", protocolError.Error()),
		})
		return
	}

	var query struct {
		RestCurrencyTime int    `json:"rest_currency_time"`
		Date             string `json:"date"`
	}
	if err = json.NewDecoder(reader).Decode(&query); err != nil {
		c.JSON(http.StatusOK, CurrencyOnlineResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("GetCurrencyOnline: %v", protocolError.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, CurrencyOnlineResponse{
		Success:          true,
		RestCurrencyTime: query.RestCurrencyTime,
		FormatDateString: query.Date,
	})
}
