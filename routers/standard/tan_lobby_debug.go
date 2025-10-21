package std_api

import (
	"bunker-core/utils"
	"bunker-lite/database"
	"bunker-lite/enhance"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TanLobbyDebugRequest ..
type TanLobbyDebugRequest struct {
	FBToken       string `json:"login_token"`
	LoginResponse string `json:"login_response"`
	RaknetRand    []byte `json:"raknet_rand"`
}

// TanLobbyDebugResponse ..
type TanLobbyDebugResponse struct {
	Success         bool   `json:"success"`
	ErrorInfo       string `json:"error_info"`
	EncryptKeyBytes []byte `json:"encrypt_key_bytes"`
	DecryptKeyBytes []byte `json:"decrypt_key_bytes"`
}

func TanLobbyDebug(c *gin.Context) {
	// parse request
	var request TanLobbyDebugRequest
	var g79UserToken string

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyDebugResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyDebug: %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.FBToken, true) {
		c.JSON(http.StatusOK, TanLobbyDebugResponse{
			ErrorInfo: "TanLobbyDebug: Invalid token was provided",
			Success:   false,
		})
		return
	}

	resp, err := enhance.ParseHttpEncrypt(request.LoginResponse)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyDebugResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyDebug: %v", err),
			Success:   false,
		})
		return
	}

	entity, ok := resp["entity"].(map[string]any)
	if ok {
		g79UserToken, ok = entity["token"].(string)
	}
	if !ok {
		c.JSON(http.StatusOK, TanLobbyDebugResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyDebug: Bad pe auth login response; resp = %#v", resp),
			Success:   false,
		})
		return
	}

	encryptedUserToken := utils.MD5Sum([]byte(g79UserToken))
	c.JSON(
		http.StatusOK,
		TanLobbyDebugResponse{
			Success:         true,
			EncryptKeyBytes: []byte(string(encryptedUserToken) + string(request.RaknetRand)),
			DecryptKeyBytes: []byte(string(request.RaknetRand) + string(encryptedUserToken)),
		},
	)
}
