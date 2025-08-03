package utils

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var (
	sessions *cache.Cache // cache[string]*sync.Map
)

func init() {
	sessions = cache.New(5*time.Minute, 5*time.Minute)
}

func GetSessionByBearer(c *gin.Context) *sync.Map {
	bearer := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
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
