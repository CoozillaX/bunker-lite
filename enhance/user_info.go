package enhance

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetName(gu *g79.G79User) (name string, err error) {
	reqBody, _ := json.Marshal(map[string]string{
		"entity_id": gu.EntityID,
	})
	reader, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79Servers.Load().CoreServerUrl + "/pe-user-detail/get").
		SetRawBody(reqBody).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		return "", fmt.Errorf("GetName: %v", protocolError.Error())
	}

	var query struct {
		Entity *struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return "", fmt.Errorf("GetName: %v", err)
	}

	return query.Entity.Name, nil
}

func ChangeName(gu *g79.G79User, userNewName string) *defines.ProtocolError {
	reqBody, _ := json.Marshal(map[string]any{
		"name": userNewName,
	})
	_, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/pe-nickname-setting/update").
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
		SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/pe-get-grow-lv-exp").
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
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return 0, 0, 0, &defines.ProtocolError{
			Message: fmt.Sprintf("GetLauncherLevel: %v", err),
		}
	}

	return query.Entity.Level, query.Entity.Exp, query.Entity.NeedExp, nil
}
