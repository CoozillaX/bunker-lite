package std_api

import (
	"bunker-lite/database"
	"encoding/json"
	"fmt"
	"net/http"

	"bunker-core/protocol/gameinfo"

	"github.com/gin-gonic/gin"
)

// TanLobbyTransferServersRequest ..
type TanLobbyTransferServersRequest struct {
	FBToken string `json:"login_token"`
}

// TanLobbyTransferServersResponse ..
type TanLobbyTransferServersResponse struct {
	Success          bool     `json:"success"`
	ErrorInfo        string   `json:"error_info"`
	RaknetServers    []string `json:"raknet_servers"`
	WebsocketServers []string `json:"websocket_servers"`
}

func TanLobbyTransferServer(c *gin.Context) {
	var request TanLobbyTransferServersRequest
	var response TanLobbyTransferServersResponse

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyTransferServersResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: %v", err),
		})
		return
	}

	if !database.CheckAuthHelperByToken(request.FBToken, true) {
		c.JSON(http.StatusOK, TanLobbyTransferServersResponse{
			Success:   false,
			ErrorInfo: "TanLobbyTransferServer: Invalid token was provided",
		})
		return
	}

	resp, err := http.Get(gameinfo.G79ServerList.TransferServerUrl)
	if err != nil {
		c.JSON(http.StatusOK, TanLobbyTransferServersResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: %v", err),
		})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, TanLobbyTransferServersResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: API Server return a non-OK status code which is %d", resp.StatusCode),
		})
		return
	}

	var serverList []struct {
		Status         int    `json:"status"`
		IP             string `json:"ip"`
		SignalWebPort  int    `json:"SignalWebPort"`
		WebsocketPorts []int  `json:"ports"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&serverList); err != nil {
		c.JSON(http.StatusOK, TanLobbyTransferServersResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("TanLobbyTransferServer: %v", err),
		})
		return
	}
	_ = resp.Body.Close()

	for _, value := range serverList {
		for _, val := range value.WebsocketPorts {
			response.RaknetServers = append(response.RaknetServers, fmt.Sprintf("%s:%d", value.IP, val))
		}
		response.WebsocketServers = append(response.WebsocketServers, fmt.Sprintf("%s:%d", value.IP, value.SignalWebPort))
	}
	response.Success = true
	c.JSON(http.StatusOK, response)
}
