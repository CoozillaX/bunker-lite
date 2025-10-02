package enhance

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"fmt"
	"net/http"
)

func ChangeName(gu *g79.G79User, userNewName string) *defines.ProtocolError {
	reqBody, _ := json.Marshal(map[string]any{
		"name": userNewName,
	})
	_, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.WebServerUrl + "/pe-nickname-setting/update").
		SetRawBody(reqBody).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		return protocolError
	}
	return nil
}

func GetLauncherLevel(gu *g79.G79User) (level int, exp int, needExp int, protocolError *defines.ProtocolError) {
	reader, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/pe-get-grow-lv-exp").
		SetRawBody([]byte("{}")).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		return 0, 0, 0, protocolError
	}

	var query struct {
		Entity struct {
			Level   int `json:"lv"`
			Exp     int `json:"exp"`
			NeedExp int `json:"need_exp"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(reader.Bytes(), &query); err != nil {
		return 0, 0, 0, &defines.ProtocolError{
			Message: fmt.Sprintf("GetLauncherLevel: %v", err),
		}
	}

	return query.Entity.Level, query.Entity.Exp, query.Entity.NeedExp, nil
}
