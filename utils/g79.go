package utils

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-core/protocol/mpay"
	"context"
	"encoding/json"
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
	g79UserCache *cache.Cache // cache[MpayUserUid]*g79UserCacheItem
)

func init() {
	g79UserCache = cache.New(25*time.Minute, 5*time.Minute)
	g79UserCache.OnEvicted(func(uid string, value any) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		item := value.(*g79UserCacheItem)
		if item.ttl > 0 && item.gu.Update(ctx) == nil { // no need to logout if update failed
			item.ttl--
			g79UserCache.SetDefault(uid, item)
		} else {
			item.gu.Logout(ctx)
		}
	})
}

func HandleG79Login(ctx context.Context, mu mpay.MpayUser) (*g79.G79User, *defines.ProtocolError) {
	if ctx.Err() != nil {
		return nil, &defines.ProtocolError{Message: ctx.Err().Error()}
	}
	if err := mu.Initialise(); err != nil {
		return nil, err
	}
	// check cache
	if cached, ok := g79UserCache.Get(mu.GetUid() + mu.GetEngineVersion()); ok {
		item := cached.(*g79UserCacheItem)
		gu := item.gu
		// if version match?
		if gu.GetEngineVersion() == mu.GetEngineVersion() && gu.GetPatchVersion() == mu.GetPatchVersion() {
			// if still valid ?
			if _, _, protocolErr := gu.AccOnlineExp(ctx); protocolErr == nil {
				item.ttl = defaultTTL // refresh ttl
				g79UserCache.SetDefault(mu.GetUid()+mu.GetEngineVersion(), item)
				return gu, nil
			}
		}
	}
	// g79 login
	gu, protocolErr := g79.Login(ctx, mu)
	if protocolErr != nil {
		return nil, protocolErr
	}
	// cache
	g79UserCache.SetDefault(mu.GetUid()+mu.GetEngineVersion(), &g79UserCacheItem{
		gu:  gu,
		ttl: defaultTTL,
	})
	return gu, nil
}

func GetG79LauncherLevel(ctx context.Context, gu *g79.G79User) (int, *defines.ProtocolError) {
	reader, ginerr := gu.CreateHttpClient(ctx).
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79Servers.Load().ApiGatewayUrl + "/pe-get-grow-lv-exp").
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
