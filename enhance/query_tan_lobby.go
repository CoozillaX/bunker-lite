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

// TanLobbyRoomInfo ..
type TanLobbyRoomInfo struct {
	G79UserUID             uint32
	RoomOwnerID            uint32
	RoomModDisplayName     []string
	RoomModDownloadURL     []string
	RoomModEncryptKey      [][]byte
	RaknetServerAddress    string
	SignalingServerAddress string
}

func QueryTanLobbyRoomInfo(gu *g79.G79User, roomID string) (result TanLobbyRoomInfo, err error) {
	var roomTransferServerID int
	var roomModItemIDs []string

	// check provided room id is empty or not
	if len(roomID) == 0 {
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: Can not provide a room ID that is empty")
	}

	// parse user unique id (g79 user uid)
	uid, err := strconv.ParseUint(gu.EntityID, 10, 32)
	if err != nil {
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", err)
	}
	result.G79UserUID = uint32(uid)

	// get room info
	reqBody, _ := json.Marshal(map[string]any{
		"name": roomID,
		"uid":  uid,
	})
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.TransferServerNewHttpUrl + "/room-with-name").
		SetRawBody(reqBody).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolErr != nil {
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", protocolErr.Error())
	}
	var query struct {
		Code int `json:"code"`
		List []struct {
			HostID           int      `json:"hid"`
			RoomID           int      `json:"rid"`
			TransferServerID int      `json:"srv"`
			ModItemIDs       []string `json:"item_ids"`
		} `json:"list"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", err)
	}

	// get room onwer id and transfer server id
	for _, value := range query.List {
		if fmt.Sprintf("%d", value.RoomID) == roomID {
			result.RoomOwnerID = uint32(value.HostID)
			roomModItemIDs = value.ModItemIDs
			roomTransferServerID = value.TransferServerID
			break
		}
	}
	if result.RoomOwnerID == 0 {
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: 未查找到本地联机 (%v), 请在确认房间状态正常后重试", roomID)
	}

	// get room mod download url and encrypt key
	if len(roomModItemIDs) > 0 {
		roomModItemIDs, result.RoomModDisplayName, result.RoomModDownloadURL, err = GetLobbyDownloadInfoByItemIDs(gu, roomModItemIDs)
		if err != nil {
			return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", err)
		}
	}
	if len(roomModItemIDs) > 0 {
		if result.RoomModEncryptKey, err = GetLobbyItemEncryptionKeys(gu, roomModItemIDs); err != nil {
			return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", err)
		}
	}

	// query transfer server list
	resp, err := http.Get(gameinfo.G79ServerList.TransferServerUrl)
	if err != nil {
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return TanLobbyRoomInfo{}, fmt.Errorf("TanLobbyTransferServer: API Server return a non-OK status code which is %d", resp.StatusCode)
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
		return TanLobbyRoomInfo{}, fmt.Errorf("QueryTanLobbyRoomInfo: %v", err)
	}

	// ensure transfer server address
	for _, value := range serverList {
		if value.ServerID == roomTransferServerID {
			result.RaknetServerAddress = fmt.Sprintf("%s:%d", value.ServerIP, value.WebsocketPorts[rand.Intn(len(value.WebsocketPorts))])
			result.SignalingServerAddress = fmt.Sprintf("%s:%d", value.ServerIP, value.SignalWebPort)
			break
		}
	}
	if len(result.RaknetServerAddress) == 0 || len(result.SignalingServerAddress) == 0 {
		return TanLobbyRoomInfo{}, fmt.Errorf("TanLobbyTransferServer: No available transfer server was found")
	}

	// return
	return result, nil
}
