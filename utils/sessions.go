package utils

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	sessions *cache.Cache // cache[string]*sync.Map
)

func init() {
	sessions = cache.New(5*time.Minute, 5*time.Minute)
}

func GetSessionByBearer(r *http.Request) *sync.Map {
	bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if bearer == "" {
		return nil
	}
	session, ok := sessions.Get(bearer)
	if !ok {
		session = &sync.Map{}
		sessions.SetDefault(bearer, session)
	}
	return session.(*sync.Map)
}
