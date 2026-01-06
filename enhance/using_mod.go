package enhance

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	currentUsingModRawBody = `{"settings":["skin_type","skin_data","persona_data","screen_config","outfit_type","personal_open","personal_ad_open","personal_tags"]}`
)

type SkinType struct {
	Type string `json:"type"`
}

type SkinData struct {
	IsSlim     bool   `json:"is_slim"`
	ItemID     string `json:"item_id"`
	SecondType int    `json:"second_type"`
}

type ScreenConfig struct {
	ItemID        string `json:"item_id"`
	OutfitLevel   *int   `json:"outfit_level,omitempty"`
	BehaviourUUID string `json:"behaviour_uuid"`
	EffectMtypeid int    `json:"effect_mtypeid"`
	EffectStypeid int    `json:"effect_stypeid"`
}

type UsingMod struct {
	SkinType         SkinType                 `json:"skin_type"`
	SkinData         SkinData                 `json:"skin_data"`
	ScreenConfig     map[string]*ScreenConfig `json:"screen_config"`
	SkinDownloadInfo *DownloadInfo
}

func GetCurrentUsingMod(gu *g79.G79User) (UsingMod, *defines.ProtocolError) {
	// 1. Do req
	reader, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/pe-get-user-setting-list").
		SetRawBody([]byte(currentUsingModRawBody)).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		return UsingMod{}, protocolError
	}

	// 2. Parse response
	var query struct {
		UsingMod UsingMod `json:"entity"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return UsingMod{}, &defines.ProtocolError{
			Message: fmt.Sprintf("GetCurrentUsingMod: %v", err),
		}
	}

	// 3. Get skin download info
	if !strings.HasPrefix(query.UsingMod.SkinData.ItemID, "-") {
		query.UsingMod.SkinDownloadInfo, protocolError = GetDownloadInfoByItemID(gu, query.UsingMod.SkinData.ItemID)
		if protocolError != nil {
			return UsingMod{}, protocolError
		}
		skinIsSlim, protocolError := GetSkinIsSlim(gu, query.UsingMod.SkinData.ItemID)
		if protocolError != nil {
			return UsingMod{}, protocolError
		}
		query.UsingMod.SkinData.IsSlim = skinIsSlim
	} else {
		query.UsingMod.SkinDownloadInfo = &DownloadInfo{
			EntityID: query.UsingMod.SkinData.ItemID,
			ResUrl:   "",
		}
	}

	return query.UsingMod, nil
}

func GetSkinIsSlim(gu *g79.G79User, itemID string) (isSlim bool, protocolError *defines.ProtocolError) {
	// 1. Make request
	var req struct {
		ItemIDList []string `json:"item_id_list"`
	}
	req.ItemIDList = []string{itemID}
	reqBody, _ := json.Marshal(req)

	// 2. Do request
	reader, protocolError := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/pe-item/query/search-by-id-list").
		SetRawBody([]byte(reqBody)).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolError != nil {
		return false, protocolError
	}

	// 3. Parse response
	var query struct {
		Entities []struct {
			SkinBodyType int `json:"skin_body_type"`
		} `json:"entities"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return false, &defines.ProtocolError{
			Message: fmt.Sprintf("GetSkinIsSlim: %v", err),
		}
	}

	// 4. Return result
	if len(query.Entities) == 1 && query.Entities[0].SkinBodyType == 1 {
		return true, nil
	}
	return false, nil
}

type PhoenixSkinInfo struct {
	EntityID string `json:"entity_id"`
	ResUrl   string `json:"res_url"`
	IsSlim   bool   `json:"is_slim"`
}

func (u UsingMod) AsPhoenixBotSkin() PhoenixSkinInfo {
	return PhoenixSkinInfo{
		EntityID: u.SkinDownloadInfo.EntityID,
		ResUrl:   u.SkinDownloadInfo.ResUrl,
		IsSlim:   u.SkinData.IsSlim,
	}
}

func (u UsingMod) AsPhoenixBotComponent() (ret map[string]*int) {
	ret = make(map[string]*int)
	for _, v := range u.ScreenConfig {
		if v.OutfitLevel == nil {
			ret[v.BehaviourUUID] = nil
			continue
		}
		var gameOutfitLevel int
		switch *v.OutfitLevel {
		case 0:
			gameOutfitLevel = 2
		case 1:
			gameOutfitLevel = 1
		case 2:
			gameOutfitLevel = 0
		}
		ret[v.BehaviourUUID] = &gameOutfitLevel
	}
	return
}
