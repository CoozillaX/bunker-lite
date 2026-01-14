package api

import (
	"bunker-lite/utils"
	"net/http"
	"strings"
	"time"

	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/mpay"
	"bunker-core/protocol/mpay/android"

	"encoding/json"

	"github.com/patrickmn/go-cache"
)

type LoginRequest struct {
	FBToken         string `json:"login_token"`
	UserName        string `json:"username"`
	Password        string `json:"password"`
	ServerCode      string `json:"server_code"`
	ServerPasscode  string `json:"server_passcode"`
	ClientPublicKey string `json:"client_public_key"`
}

type LoginResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ChainInfo   string `json:"chainInfo,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	GrowthLevel int    `json:"growth_level,omitempty"`
	Token       string `json:"token,omitempty"`
}

var versionCache = cache.New(24*time.Hour, time.Hour) // cache[serverCode]bedrockVersion

func requestServerInfo(
	mu mpay.MpayUser,
	req *LoginRequest,
) (*g79.G79User, *g79.RentalServerInfo, *defines.ProtocolError) {
	// change engine version by cache
	if value, ok := versionCache.Get(req.ServerCode); ok {
		if err := mu.UpdateGameInfoByBedrockVersion(value.(string)); err != nil {
			return nil, nil, &defines.ProtocolError{Message: err.Error()}
		}
	}
	// g79 login
	gu, protocolErr := utils.HandleG79Login(mu)
	if protocolErr != nil {
		return nil, nil, protocolErr
	}
	// chain info
	rentalInfo, protocolErr := gu.ImpactRentalServer(req.ServerCode, req.ServerPasscode, req.ClientPublicKey)
	if protocolErr != nil {
		return nil, nil, protocolErr
	}
	// cache version
	rentalBedrockVersion := strings.TrimSuffix(rentalInfo.MCVersion, "-release")
	versionCache.SetDefault(req.ServerCode, rentalBedrockVersion)
	// check version
	if gu.GetBedrockVersion() != rentalBedrockVersion {
		// re-login and get chain with updated engine version
		return requestServerInfo(mu, req)
	}
	return gu, rentalInfo, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// parse request
	var req LoginRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// parse token
	var mu *android.AndroidMpayUser
	if req.FBToken != "" {
		mu, _ = utils.DecodeFBToken(req.FBToken)
	}
	if mu == nil {
		mu = &android.AndroidMpayUser{}
	}
	// try mpay login if no token
	if mu.MpayToken == "" {
		if protocolErr := mu.GuestLogin(); protocolErr != nil {
			json.NewEncoder(w).Encode(&LoginResponse{
				Success: false,
				Message: protocolErr.Error(),
				Token:   utils.EncodeFBToken(mu),
			})
			return
		}
	}
	// dry login, should always do dry login first to initialise mpay user
	if req.ServerCode == "::DRY::" && req.ServerPasscode == "::DRY::" {
		json.NewEncoder(w).Encode(&LoginResponse{
			Success: true,
			Message: "ok",
			Token:   utils.EncodeFBToken(mu),
		})
		return
	}
	// g79 login and request server info
	gu, serverInfo, protocolErr := requestServerInfo(mu, &req)
	if protocolErr != nil {
		json.NewEncoder(w).Encode(&LoginResponse{
			Success: false,
			Message: protocolErr.Error(),
		})
		return
	}
	// fetch growth level
	growthLevel, _ := utils.GetG79LauncherLevel(gu)
	// save info for anti-cheat callback
	session := utils.GetSessionByBearer(r)
	session.Store(session_key_entity_id, gu.EntityID)
	session.Store(session_key_engine_version, gu.GetEngineVersion())
	session.Store(session_key_patch_version, gu.GetPatchVersion())
	session.Store(session_key_platform, gu.GetSystemName())
	// response
	json.NewEncoder(w).Encode(&LoginResponse{
		Success:     true,
		Message:     "ok",
		ChainInfo:   serverInfo.ChainInfo,
		IPAddress:   serverInfo.IPAddress,
		GrowthLevel: growthLevel,
	})
}
