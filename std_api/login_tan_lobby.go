package std_api

import (
	"bunker-core/utils"
	"bunker-lite/database"
	"bunker-lite/define"
	cryptoRand "crypto/rand"
	"fmt"
	mathRand "math/rand"
	"net/http"
	"strconv"

	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

// TanLobbyLoginRequest ..
type TanLobbyLoginRequest struct {
	FBToken string `json:"login_token"`
	RoomID  string `json:"room_id"`
}

// TanLobbyLoginResponse ..
type TanLobbyLoginResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	RoomOwnerID    uint32 `json:"room_owner_id"`
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

func TanLobbyLogin(c *gin.Context) {
	// parse request
	var request TanLobbyLoginRequest
	var helper define.AuthServerHelper

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.FBToken, true) {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: "TanLobbyLogin: Invalid token was provided",
		})
		return
	}
	helper = database.GetAuthHelperByToken(request.FBToken, true)

	// decode to mpay user
	var mu *defines.MpayUser = new(defines.MpayUser)
	if err = json.Unmarshal(helper.MpayUserData, mu); err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
		})
		return
	}

	// g79 login
	gu, protocolErr := g79.Login(gameinfo.DefaultEngineVersion, mu)
	if protocolErr != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", protocolErr.Error()),
		})
		return
	}

	// parse user unique id (g79 user uid)
	g79UserUID, err := strconv.ParseUint(gu.EntityID, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
		})
		return
	}

	// get room info
	reqBody, _ := json.Marshal(map[string]any{
		"name": request.RoomID,
		"uid":  g79UserUID,
	})
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.TransferServerNewHttpUrl + "/room-with-name").
		SetRawBody(reqBody).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolErr != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", protocolErr.Error()),
		})
		return
	}
	var query struct {
		Code int `json:"code"`
		List []struct {
			HostID           int `json:"hid"`
			RoomID           int `json:"rid"`
			TransferServerID int `json:"srv"`
		} `json:"list"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
		})
		return
	}

	// get room onwer id and transfer server id
	var roomOwnerID uint32
	var roomTransferServerID int
	for _, value := range query.List {
		if fmt.Sprintf("%d", value.RoomID) == request.RoomID {
			roomOwnerID = uint32(value.HostID)
			roomTransferServerID = value.TransferServerID
			break
		}
	}
	if roomOwnerID == 0 {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: 未查找到本地联机 (%v), 请在确认房间状态正常后重试", request.RoomID),
		})
		return
	}

	// query transfer server list
	resp, err := http.Get(gameinfo.G79ServerList.TransferServerUrl)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: %v", err),
		})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: API Server return a non-OK status code which is %d", resp.StatusCode),
		})
		return
	}

	// parse transfer server list
	var serverList []struct {
		Status         int    `json:"status"`
		ServerIP       string `json:"ip"`
		ServerID       int    `json:"id"`
		SignalWebPort  int    `json:"SignalWebPort"`
		WebsocketPorts []int  `json:"ports"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&serverList); err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: %v", err),
		})
		return
	}
	_ = resp.Body.Close()

	// ensure transfer server address
	var raknetServerAddress string
	var signalingServerAddress string
	for _, value := range serverList {
		if value.ServerID == roomTransferServerID {
			raknetServerAddress = fmt.Sprintf("%s:%d", value.ServerIP, value.WebsocketPorts[mathRand.Intn(len(value.WebsocketPorts))])
			signalingServerAddress = fmt.Sprintf("%s:%d", value.ServerIP, value.SignalWebPort)
			break
		}
	}
	if len(raknetServerAddress) == 0 || len(signalingServerAddress) == 0 {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: "TanLobbyTransferServer: No available transfer server was found",
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
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
		})
		return
	}
	raknetAESRand = raknetAESRand[0:16]

	// compute signaling ticket by g79 token and seed
	signalingTicket, err := utils.AES_ECB_PKCS7Encrypt([]byte(gu.G79Token), signalingSeed)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyLoginResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyLogin: %v", err),
		})
		return
	}
	signalingTicket = signalingTicket[0:16]

	c.JSON(
		http.StatusOK,
		TanLobbyLoginResponse{
			Success:                true,
			RoomOwnerID:            roomOwnerID,
			UserUniqueID:           uint32(g79UserUID),
			UserPlayerName:         gu.Username,
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
