package enhance

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
)

func SelectTransferServer(gu *g79.G79User) (g79UserUID uint32, raknetServerAddress string, signalingServerAddress string, err error) {
	// parse user unique id (g79 user uid)
	uid, err := strconv.ParseUint(gu.EntityID, 10, 32)
	if err != nil {
		return 0, "", "", fmt.Errorf("SelectTransferServer: %v", err)
	}
	g79UserUID = uint32(uid)

	// query transfer server list
	resp, err := http.Get(gameinfo.G79Servers.Load().TransferServerUrl)
	if err != nil {
		return 0, "", "", fmt.Errorf("SelectTransferServer: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, "", "", fmt.Errorf("SelectTransferServer: API Server return a non-OK status code which is %d", resp.StatusCode)
	}

	// parse transfer server list
	var serverList []struct {
		Status         int    `json:"status"`
		ServerIP       string `json:"ip"`
		ServerID       int    `json:"id"`
		SignalWebPort  int    `json:"SignalWebPort"`
		WebsocketPorts []int  `json:"ports"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&serverList); err != nil {
		return 0, "", "", fmt.Errorf("SelectTransferServer: %v", err)
	}

	// ensure transfer server address
	server := serverList[rand.Intn(len(serverList))]
	raknetServerAddress = fmt.Sprintf("%s:%d", server.ServerIP, server.WebsocketPorts[rand.Intn(len(server.WebsocketPorts))])
	signalingServerAddress = fmt.Sprintf("%s:%d", server.ServerIP, server.SignalWebPort)

	// return
	return
}
