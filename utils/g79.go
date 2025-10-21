package utils

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

var g79UserCache *cache.Cache // cache[MpayUserUid]*g79.G79User

func init() {
	g79UserCache = cache.New(25*time.Minute, 5*time.Minute)
	g79UserCache.OnEvicted(func(uid string, value any) {
		gu := value.(*g79.G79User)
		gu.AccOnlineExp()
	})
}

func HandleG79Login(engineVersion string, mu *defines.MpayUser) (*g79.G79User, *defines.ProtocolError) {
	// check cache
	if cached, ok := g79UserCache.Get(mu.Uid); ok {
		gu := cached.(*g79.G79User)
		// if version match?
		if gu.GameInfo.EngineVersion == engineVersion {
			// if expired?
			if ginerr := gu.Update(); ginerr == nil {
				gu.AccOnlineExp()
				g79UserCache.SetDefault(mu.Uid, gu)
				return gu, nil
			}
		}
	}
	// g79 login
	gu, protocolErr := g79.Login(engineVersion, mu)
	if protocolErr != nil {
		return nil, protocolErr
	}
	// cache
	g79UserCache.SetDefault(mu.Uid, gu)
	return gu, nil
}

func GetG79LauncherLevel(gu *g79.G79User) (int, *defines.ProtocolError) {
	reader, ginerr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/pe-get-grow-lv-exp").
		SetRawBody([]byte("{}")).
		SetTokenMode(g79.TOKEN_MODE_NORMAL).
		Do()
	if ginerr != nil {
		return 0, ginerr
	}
	var query struct {
		Entity struct {
			Level int `json:"lv"`
		} `json:"entity"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return 0, &defines.ProtocolError{Message: err.Error()}
	}
	return query.Entity.Level, nil
}
