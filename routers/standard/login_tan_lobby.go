package std_api

import (
	"bunker-core/protocol/gameinfo"
	"bunker-core/utils"
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/enhance"
	cryptoRand "crypto/rand"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TanLobbyLoginRequest ..
type TanLobbyLoginRequest struct {
	FBToken            string `json:"login_token"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data"`
	RoomID             string `json:"room_id"`
}

// TanLobbyLoginResponse ..
type TanLobbyLoginResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	UserUniqueID   uint32                  `json:"user_unique_id"`
	UserPlayerName string                  `json:"user_player_name"`
	BotLevel       int                     `json:"growth_level"`
	BotSkin        enhance.PhoenixSkinInfo `json:"skin_info"`
	BotComponent   map[string]*int         `json:"outfit_info,omitempty"`

	RoomOwnerID        uint32   `json:"room_owner_id"`
	RoomModDisplayName []string `json:"room_mod_display_name"`
	RoomModDownloadURL []string `json:"room_mod_download_url"`
	RoomModEncryptKey  [][]byte `json:"room_mod_encrypt_key"`

	RaknetServerAddress string `json:"raknet_server_address"`
	RaknetRand          []byte `json:"raknet_rand"`
	RaknetAESRand       []byte `json:"raknet_aes_rand"`
	EncryptKeyBytes     []byte `json:"encrypt_key_bytes"`
	DecryptKeyBytes     []byte `json:"decrypt_key_bytes"`

	SignalingServerAddress string `json:"signaling_server_address"`
	SignalingSeed          []byte `json:"signaling_seed"`
	SignalingTicket        []byte `json:"signaling_ticket"`
}

func TanLobbyLogin(c *gin.Context) {
	// parse request
	var request TanLobbyLoginRequest
	var helper define.AuthServerHelper

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.FBToken, true) {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: "TanLobbyLogin: Invalid token was provided",
			Success:   false,
		})
		return
	}
	helper = database.GetAuthHelperByToken(request.FBToken, true)

	// g79 login
	gu, _, err := database.LoadOrRegisterActiveG79User(
		helper,
		gameinfo.DefaultEngineVersion,
		request.ProvidedPeAuthData,
		request.ProvidedSaAuthData,
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
			Success:   false,
		})
		return
	}

	// query tan lobby room info
	roomInfo, err := enhance.QueryTanLobbyRoomInfo(gu, request.RoomID)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
			Success:   false,
		})
		return
	}

	// Get launcher level and current using mod
	launcherLevel, _, _, protocolError := enhance.GetLauncherLevel(gu)
	if protocolError != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", protocolError.Error()),
			Success:   false,
		})
		return
	}
	currentUsingMod, protocolError := enhance.GetCurrentUsingMod(gu)
	if protocolError != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", protocolError.Error()),
			Success:   false,
		})
		return
	}

	// generate rand and seed
	raknetRand := make([]byte, 16)
	signalingSeed := make([]byte, 16)
	_, _ = cryptoRand.Read(raknetRand)
	_, _ = cryptoRand.Read(signalingSeed)

	// compute encrypted token and key to encrypt/decrypt raknet session
	encryptedUserToken := utils.MD5Sum([]byte(gu.G79Token))
	encryptKeyBytes := []byte(string(encryptedUserToken) + string(raknetRand))
	decryptKeyBytes := []byte(string(raknetRand) + string(encryptedUserToken))

	// compute raknet aes rand
	raknetAESRand, err := utils.AES_ECB_PKCS7Encrypt(encryptedUserToken, raknetRand)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
			Success:   false,
		})
		return
	}
	raknetAESRand = raknetAESRand[0:16]

	// compute signaling ticket by g79 token and seed
	signalingTicket, err := utils.AES_ECB_PKCS7Encrypt([]byte(gu.G79Token), signalingSeed)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
			Success:   false,
		})
		return
	}
	signalingTicket = signalingTicket[0:16]

	c.JSON(
		http.StatusOK,
		TanLobbyLoginResponse{
			Success:                true,
			UserUniqueID:           roomInfo.G79UserUID,
			UserPlayerName:         gu.Username,
			BotLevel:               launcherLevel,
			BotSkin:                currentUsingMod.AsPhoenixBotSkin(),
			BotComponent:           currentUsingMod.AsPhoenixBotComponent(),
			RoomOwnerID:            roomInfo.RoomOwnerID,
			RoomModDisplayName:     roomInfo.RoomModDisplayName,
			RoomModDownloadURL:     roomInfo.RoomModDownloadURL,
			RoomModEncryptKey:      roomInfo.RoomModEncryptKey,
			RaknetServerAddress:    roomInfo.RaknetServerAddress,
			RaknetRand:             raknetRand,
			RaknetAESRand:          raknetAESRand,
			EncryptKeyBytes:        encryptKeyBytes,
			DecryptKeyBytes:        decryptKeyBytes,
			SignalingServerAddress: roomInfo.SignalingServerAddress,
			SignalingSeed:          signalingSeed,
			SignalingTicket:        signalingTicket,
		},
	)
}
