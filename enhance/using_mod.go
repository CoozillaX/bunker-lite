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
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.WebServerUrl + "/pe-get-user-setting-list").
		SetRawBody([]byte(currentUsingModRawBody)).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if protocolErr != nil {
		return UsingMod{}, protocolErr
	}

	// 2. Parse response
	var query struct {
		UsingMod UsingMod `json:"entity"`
	}
	if err := json.Unmarshal(reader.Bytes(), &query); err != nil {
		return UsingMod{}, &defines.ProtocolError{
			Message: fmt.Sprintf("GetCurrentUsingMod: %v", err),
		}
	}

	// 3. Get skin download info
	if !strings.HasPrefix(query.UsingMod.SkinData.ItemID, "-") {
		query.UsingMod.SkinDownloadInfo, protocolErr = GetDownloadInfoByItemID(gu, query.UsingMod.SkinData.ItemID)
		if protocolErr != nil {
			return UsingMod{}, protocolErr
		}
	} else {
		query.UsingMod.SkinDownloadInfo = &DownloadInfo{
			EntityID: query.UsingMod.SkinData.ItemID,
			ResUrl:   "",
		}
	}

	return query.UsingMod, nil
}

type PhoenixEnhanceInfo struct {
	SkinInfo   PhoenixSkinInfo `json:"skin_info"`
	OutfitInfo map[string]*int `json:"outfit_info,omitempty"`
}

type PhoenixSkinInfo struct {
	EntityID string `json:"entity_id"`
	ResUrl   string `json:"res_url"`
	IsSlim   bool   `json:"is_slim"`
}

func (u UsingMod) AsPhoenixEnhanceInfo() PhoenixEnhanceInfo {
	return PhoenixEnhanceInfo{
		SkinInfo: PhoenixSkinInfo{
			EntityID: u.SkinDownloadInfo.EntityID,
			ResUrl:   u.SkinDownloadInfo.ResUrl,
			IsSlim:   u.SkinData.IsSlim,
		},
		OutfitInfo: u.GetConfigUUID2OutfitLevel(),
	}
}

func (u UsingMod) GetConfigUUID2OutfitLevel() (ret map[string]*int) {
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
