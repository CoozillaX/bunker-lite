package enhance

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-core/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func GetLobbyDownloadInfoByItemIDs(gu *g79.G79User, itemIDs []string) (
	realItemIds []string,
	modDisplayName []string,
	modDownloadURL []string,
	err error,
) {
	// 1. Do req
	reqBody, _ := json.Marshal(map[string]any{
		"item_ids": itemIDs,
	})
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.WebServerUrl + "/pe-item/query/search-lobby-by-id-list").
		SetRawBody(reqBody).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolErr != nil {
		return nil, nil, nil, fmt.Errorf("GetLobbyDownloadInfoByItemIDs: %v", protocolErr.Error())
	}

	// 2. Parse response
	var query struct {
		Entities []struct {
			ModItemID      string `json:"iid"`
			ModDisplayName string `json:"res_name"`
			LobbyResUrl    string `json:"lobby_res_url"`
		} `json:"entities"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return nil, nil, nil, fmt.Errorf("GetLobbyDownloadInfoByItemIDs: %v", err)
	}

	// 3. Get urls
	for _, item := range query.Entities {
		if len(item.LobbyResUrl) > 0 {
			modDisplayName = append(modDisplayName, item.ModDisplayName)
			modDownloadURL = append(modDownloadURL, item.LobbyResUrl)
			realItemIds = append(realItemIds, item.ModItemID)
		}
	}

	return
}

func GetLobbyItemEncryptionKeys(gu *g79.G79User, itemIDs []string) (result [][]byte, err error) {
	// 1. Do req
	reqBody, _ := json.Marshal(map[string]any{
		"device_id": gu.MpayUser.UrsUdid,
		"item_ids":  itemIDs,
	})
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.CoreServerUrl + "/pe-item/get-encryption-key-list-for-guests").
		SetRawBody(reqBody).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		SetEncryptSuffix(0xc).
		Do()
	if protocolErr != nil {
		return nil, fmt.Errorf("GetLobbyItemEncryptionKeys: %v", protocolErr.Error())
	}

	// 2. Parse response
	var query struct {
		Entities []struct {
			ModItemID  string `json:"item_id"`
			DecryptJWT string `json:"jwt"`
		} `json:"entities"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return nil, fmt.Errorf("GetLobbyItemEncryptionKeys: %v", err)
	}

	// 3. Get decrypt key
	mapping := make(map[string][]byte)
	for _, item := range query.Entities {
		// 3.1. No jwt, so set a nil key
		if len(item.DecryptJWT) == 0 {
			mapping[item.ModItemID] = make([]byte, 0)
			continue
		}
		// 3.2. Decrypt jwt
		token, _, err := new(jwt.Parser).ParseUnverified(item.DecryptJWT, jwt.MapClaims{})
		if err != nil {
			return nil, fmt.Errorf("GetLobbyItemEncryptionKeys: %v", err)
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("GetLobbyItemEncryptionKeys: Invalid jwt (should never happened) (stage 0)")
		}
		// 3.3 Get encrypt key
		contentKey, ok := claims["contentKey"].(string)
		if !ok {
			return nil, fmt.Errorf("GetLobbyItemEncryptionKeys: Invalid jwt (should never happened) (stage 1)")
		}
		mapping[item.ModItemID] = utils.GetRecordEncryptKey(contentKey, gu.EntityID, gu.MpayUser.UrsUdid)
	}
	for _, value := range itemIDs {
		result = append(result, mapping[value])
	}

	return
}
