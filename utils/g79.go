package utils

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

const defaultTTL = 4

type g79UserCacheItem struct {
	gu  *g79.G79User
	ttl int
}

var (
	g79UserCache       *cache.Cache // cache[MpayUserUid]*g79UserCacheItem
	g79UserCacheLogger = log.Default()
)

func init() {
	g79UserCacheLogger.SetPrefix("[G79UserCache] ")
	g79UserCache = cache.New(25*time.Minute, 5*time.Minute)
	g79UserCache.OnEvicted(func(uid string, value any) {
		item := value.(*g79UserCacheItem)
		if item.ttl > 0 && item.gu.Update() == nil { // no need to logout if update failed
			item.ttl--
			g79UserCache.SetDefault(uid, item)
			g79UserCacheLogger.Printf("CACHE REFRESH: uid=%s, engineVersion=%s, new ttl=%d\n", uid, item.gu.GameInfo.EngineVersion, item.ttl)
		} else {
			item.gu.Logout()
		}
	})
}

func HandleG79Login(engineVersion string, mu *defines.MpayUser) (*g79.G79User, *defines.ProtocolError) {
	// check cache
	if cached, ok := g79UserCache.Get(mu.Uid); ok {
		item := cached.(*g79UserCacheItem)
		gu := item.gu
		// if version match? and if still valid?
		if gu.GameInfo.EngineVersion == engineVersion && gu.AccOnlineExp() == nil {
			g79UserCacheLogger.Printf("CACHE HIT: uid=%s, engineVersion=%s, old ttl=%d\n", mu.Uid, engineVersion, item.ttl)
			item.ttl = defaultTTL // refresh ttl
			g79UserCache.SetDefault(mu.Uid, item)
			return gu, nil
		}
	}
	// g79 login
	gu, protocolErr := g79.Login(engineVersion, mu)
	if protocolErr != nil {
		return nil, protocolErr
	}
	g79UserCacheLogger.Printf("NEW LOGIN: uid=%s, engineVersion=%s\n", mu.Uid, engineVersion)
	// cache
	g79UserCache.SetDefault(mu.Uid, &g79UserCacheItem{
		gu:  gu,
		ttl: defaultTTL,
	})
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
