package enhance

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"fmt"
	"net/http"
)

type DownloadInfo struct {
	EntityID string `json:"entity_id"`
	ResUrl   string `json:"res_url"`
}

func GetDownloadInfoByItemID(gu *g79.G79User, id string) (*DownloadInfo, *defines.ProtocolError) {
	// 1. Do req
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/pe-download-item/get-download-info").
		SetRawBody(fmt.Appendf(nil, `{"item_id":"%s"}`, id)).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolErr != nil {
		return nil, protocolErr
	}

	// 2. Parse response
	var query struct {
		Entity DownloadInfo `json:"entity"`
	}
	if err := json.Unmarshal(reader.Bytes(), &query); err != nil {
		return nil, &defines.ProtocolError{
			Message: fmt.Sprintf("GetDownloadInfoByItemID: %v", err),
		}
	}

	return &query.Entity, nil
}
