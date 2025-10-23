package std_api

import (
	"bunker-core/utils"
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/enhance"
	cryptoRand "crypto/rand"
	"fmt"
	"net/http"

	"bunker-core/protocol/gameinfo"

	"github.com/gin-gonic/gin"
)

// TanLobbyCreateRequest ..
type TanLobbyCreateRequest struct {
	FBToken            string `json:"login_token"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data"`
}

// TanLobbyCreateResponse ..
type TanLobbyCreateResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	UserUniqueID   uint32 `json:"user_unique_id"`
	UserPlayerName string `json:"user_player_name"`

	RaknetServerAddress string `json:"raknet_server_address"`
	RaknetRand          []byte `json:"raknet_rand"`
	RaknetAESRand       []byte `json:"raknet_aes_rand"`
	EncryptKeyBytes     []byte `json:"encrypt_key_bytes"`
	DecryptKeyBytes     []byte `json:"decrypt_key_bytes"`

	SignalingServerAddress string `json:"signaling_server_address"`
	SignalingSeed          []byte `json:"signaling_seed"`
	SignalingTicket        []byte `json:"signaling_ticket"`
}

func TanLobbyCreate(c *gin.Context) {
	// parse request
	var request TanLobbyCreateRequest
	var helper define.AuthServerHelper

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyCreate: %v", err),
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.FBToken, true) {
		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:   false,
			ErrorInfo: "TanLobbyCreate: Invalid token was provided",
		})
		return
	}
	helper = database.GetAuthHelperByToken(request.FBToken, true)

	// g79 login
	activeGu, err := database.LoadOrRegisterActiveG79User(
		helper,
		gameinfo.DefaultEngineVersion,
		request.ProvidedPeAuthData,
		request.ProvidedSaAuthData,
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyCreate: %v", err),
		})
		return
	}

	// select transfer server
	g79UserUID, raknetServerAddress, signalingServerAddress, err := enhance.SelectTransferServer(activeGu.RecordG79UserData)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyCreate: %v", err),
		})
		return
	}

	// generate rand and seed
	raknetRand := make([]byte, 16)
	signalingSeed := make([]byte, 16)
	_, _ = cryptoRand.Read(raknetRand)
	_, _ = cryptoRand.Read(signalingSeed)

	// compute encrypted token and key to encrypt/decrypt raknet session
	encryptedUserToken := utils.MD5Sum([]byte(activeGu.RecordG79UserData.G79Token))
	encryptKeyBytes := []byte(string(encryptedUserToken) + string(raknetRand))
	decryptKeyBytes := []byte(string(raknetRand) + string(encryptedUserToken))

	// compute raknet aes rand
	raknetAESRand, err := utils.AES_ECB_PKCS7Encrypt(encryptedUserToken, raknetRand)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyCreate: %v", err),
		})
		return
	}
	raknetAESRand = raknetAESRand[0:16]

	// compute signaling ticket by g79 token and seed
	signalingTicket, err := utils.AES_ECB_PKCS7Encrypt([]byte(activeGu.RecordG79UserData.G79Token), signalingSeed)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyCreateResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyCreate: %v", err),
		})
		return
	}
	signalingTicket = signalingTicket[0:16]

	c.JSON(
		http.StatusOK,
		TanLobbyCreateResponse{
			Success:                true,
			UserUniqueID:           g79UserUID,
			UserPlayerName:         activeGu.RecordG79UserData.Username,
			RaknetServerAddress:    raknetServerAddress,
			RaknetRand:             raknetRand,
			RaknetAESRand:          raknetAESRand,
			EncryptKeyBytes:        encryptKeyBytes,
			DecryptKeyBytes:        decryptKeyBytes,
			SignalingServerAddress: signalingServerAddress,
			SignalingSeed:          signalingSeed,
			SignalingTicket:        signalingTicket,
		},
	)
}
