package std_api

import (
	"bunker-core/utils"
	"bunker-lite/database"
	"bunker-lite/define"
	cryptoRand "crypto/rand"
	"fmt"
	"net/http"

	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

// TanLobbyRefreshRequest ..
type TanLobbyRefreshRequest struct {
	FBToken string `json:"login_token"`
}

// TanLobbyRefreshResponse ..
type TanLobbyRefreshResponse struct {
	Success         bool   `json:"success"`
	ErrorInfo       string `json:"error_info"`
	SignalingSeed   []byte `json:"signaling_seed"`
	SignalingTicket []byte `json:"signaling_ticket"`
}

func TanLobbyRefresh(c *gin.Context) {
	// parse request
	var request TanLobbyRefreshRequest
	var helper define.AuthServerHelper

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyRefreshResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyRefresh: %v", err),
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.FBToken, true) {
		c.JSON(http.StatusOK, TanLobbyRefreshResponse{
			Success:   false,
			ErrorInfo: "TanLobbyRefresh: Invalid token was provided",
		})
		return
	}
	helper = database.GetAuthHelperByToken(request.FBToken, true)

	// decode to mpay user
	var mu *defines.MpayUser = new(defines.MpayUser)
	if err = json.Unmarshal(helper.MpayUserData, mu); err != nil {
		c.JSON(http.StatusOK, TanLobbyRefreshResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyRefresh: %v", err),
		})
		return
	}

	// g79 login
	gu, protocolErr := g79.Login(gameinfo.DefaultEngineVersion, mu)
	if protocolErr != nil {
		c.JSON(http.StatusOK, TanLobbyRefreshResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyRefresh: %v", protocolErr.Error()),
		})
		return
	}

	// generate seed
	signalingSeed := make([]byte, 16)
	_, _ = cryptoRand.Read(signalingSeed)

	// compute signaling ticket by g79 token and seed
	signalingTicket, err := utils.AES_ECB_PKCS7Encrypt([]byte(gu.G79Token), signalingSeed)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyRefreshResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyRefresh: %v", err),
		})
		return
	}
	signalingTicket = signalingTicket[0:16]

	c.JSON(
		http.StatusOK,
		TanLobbyRefreshResponse{
			Success:         true,
			SignalingSeed:   signalingSeed,
			SignalingTicket: signalingTicket,
		},
	)
}
