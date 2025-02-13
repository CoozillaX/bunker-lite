package api

import (
	"bunker-lite/utils"
	"net/http"
	"time"

	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-core/protocol/mpay"

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
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ChainInfo string `json:"chainInfo,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Token     string `json:"token,omitempty"`
}

var versionCache = cache.New(24*time.Hour, time.Hour) // cache[serverCode]serverVersion

func requestServerInfo(
	engineVersion string,
	mu *defines.MpayUser,
	req *LoginRequest,
) (*g79.G79User, *g79.RentalServerInfo, *defines.ProtocolError) {
	// change engine version by cache
	if value, ok := versionCache.Get(req.ServerCode); ok {
		engineVersion = value.(string)
	}
	// g79 login
	gu, protocolErr := g79.Login(engineVersion, mu)
	if protocolErr != nil {
		return nil, nil, protocolErr
	}
	// chain info
	rentalInfo, protocolErr := gu.ImpactRentalServer(req.ServerCode, req.ServerPasscode, req.ClientPublicKey)
	if protocolErr != nil {
		return nil, nil, protocolErr
	}
	// cache version
	versionCache.SetDefault(req.ServerCode, rentalInfo.MCVersion)
	// check version
	currentGameInfo, err := gameinfo.GetInfoByGameVersion(rentalInfo.MCVersion)
	if err != nil {
		return nil, nil, &defines.ProtocolError{Message: err.Error()}
	}
	if gu.GameInfo.EngineVersion != currentGameInfo.EngineVersion {
		// re-login and get chain with updated engine version
		return requestServerInfo(currentGameInfo.EngineVersion, mu, req)
	}
	return gu, rentalInfo, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	// parse request
	var req LoginRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// dry login
	if req.ServerCode == "::DRY::" && req.ServerPasscode == "::DRY::" {
		json.NewEncoder(w).Encode(&LoginResponse{
			Success: true,
			Message: "ok",
			Token:   "e30=", // empty token
		})
		return
	}
	// parse token
	var err error
	var mu *defines.MpayUser
	if req.FBToken != "" {
		mu, err = utils.DecodeFBToken(req.FBToken)
		if err != nil {
			json.NewEncoder(w).Encode(&LoginResponse{
				Success: true,
				Message: err.Error(),
			})
			return
		}
	}
	// try mpay login
	var protocolErr *defines.ProtocolError
	if mu, protocolErr = mpay.CreateLoginHelper(mu).GuestLogin(); protocolErr != nil {
		json.NewEncoder(w).Encode(&LoginResponse{
			Success: false,
			Message: protocolErr.Error(),
			Token:   utils.EncodeFBToken(mu),
		})
		return
	}
	// g79 login and request server info
	gu, serverInfo, protocolErr := requestServerInfo(gameinfo.DefaultEngineVersion, mu, &req)
	if protocolErr != nil {
		json.NewEncoder(w).Encode(&LoginResponse{
			Success: false,
			Message: protocolErr.Error(),
		})
		return
	}
	// save info for anti-cheat callback
	session := utils.GetSessionByBearer(r)
	session.Store(session_key_entity_id, gu.EntityID)
	session.Store(session_key_engine_version, gu.GameInfo.EngineVersion)
	session.Store(session_key_patch_version, gu.GameInfo.PatchVersion)
	// response
	json.NewEncoder(w).Encode(&LoginResponse{
		Success:   true,
		Message:   "success",
		ChainInfo: serverInfo.ChainInfo,
		IPAddress: serverInfo.IPAddress,
	})
}
