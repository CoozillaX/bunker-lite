package std_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/enhance"
	"bunker-lite/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// cache[serverCode]serverVersion
var versionCache = cache.New(24*time.Hour, time.Hour)

// 客户端向验证服务器发送的请求体，
// 用于获得 FBToken，
// 或使客户端登录到网易租赁服。
// AuthResponse 是该请求体对应的响应体
type AuthRequest struct {
	/*
		此字段非空时，则下方 UserName 和 Password 为空，
		否则反之。

		当 FBToken 或 UserName、Password 二者中任意一个
		填写的值正确时，用户将登录到用户中心，然后进入租赁服
	*/
	FBToken string `json:"login_token,omitempty"`
	/*
		仅赞颂者使用。

		此字段可以是 JSON 字符串，
		或者加密的字符串。
		一切取决于用户。
	*/
	ProvidedSauthData string `json:"provided_sa_auth_data"`

	UserName string `json:"username,omitempty"` // 用户在用户中心的用户名
	Password string `json:"password,omitempty"` // 用户在用户中心的密码

	ServerCode     string `json:"server_code"`     // 要进入的租赁服的 服务器号
	ServerPassword string `json:"server_passcode"` // 该租赁服的 密码

	ClientPublicKey string `json:"client_public_key"` // ...
}

// 验证服务器对 AuthRequest 的响应体
type AuthResponse struct {
	/*
		描述请求的成功状态。

		如果成功，则其余的所有非可选字段都将有值，
		这也包括 Message 本身。

		如果失败，除了本字段和 Message 以外，
		其余所有字段都为默认的零值，
		同时 Message 会展示对应的失败原因
	*/
	SuccessStates bool   `json:"success"`
	ServerMessage string `json:"server_msg,omitempty"` // 来自验证服务器的消息
	Message

	BotLevel     int                     `json:"growth_level"`          // 机器人的等级
	BotSkin      enhance.PhoenixSkinInfo `json:"skin_info,omitempty"`   // 机器人的皮肤信息
	BotComponent map[string]*int         `json:"outfit_info,omitempty"` // 机器人当前已加载的组件及其附加值

	FBToken    string `json:"token"`      // ...
	MasterName string `json:"respond_to"` // 机器人主人的游戏名称

	RentalServerIP string `json:"ip_address"` // 欲登录的租赁服的 IP 地址
	ChainInfo      string `json:"chainInfo"`  // 欲登录的租赁服的链请求
}

// 描述 AuthResponse 所附带的额外信息
type Message struct {
	/*
		若 AuthRequest 成功，
		则对于原生的 FastBuilder 的验证服务器(mv4)，
		此字段为 "正常返回"；
		否则，对于 咕咕酱及其开发团队 的验证服务器，
		此字段为 "well down"。

		当 AuthRequest 失败时，
		若此字段非空，则它将阐明对应的失败原因，
		否则，由下方的 Translation 揭示具体的原因
	*/
	Information string `json:"message,omitempty"`
	// 表示错误码，且可以与 i18n 中所记的映射对应。
	// 如果不存在，则其默认值为 0，
	// 如果未使用，则其默认值为 -1
	Translation int `json:"translation,omitempty"`
}

// requestServerInfo ..
func requestServerInfo(
	gu *g79.G79User,
	req *AuthRequest,
) (
	launcherLevel int,
	currentUsingMod enhance.UsingMod,
	rentalServerInfo *g79.RentalServerInfo,
	protocolErr *defines.ProtocolError,
) {
	// launcher level
	launcherLevel, _, _, protocolErr = enhance.GetLauncherLevel(gu)
	if protocolErr != nil {
		return 0, enhance.UsingMod{}, nil, protocolErr
	}
	// using mod
	currentUsingMod, protocolErr = enhance.GetCurrentUsingMod(gu)
	if protocolErr != nil {
		return 0, enhance.UsingMod{}, nil, protocolErr
	}
	// chain info
	rentalInfo, protocolErr := gu.ImpactRentalServer(req.ServerCode, req.ServerCode, req.ClientPublicKey)
	if protocolErr != nil {
		return 0, enhance.UsingMod{}, nil, protocolErr
	}
	// cache version
	currentGameInfo, err := gameinfo.GetInfoByGameVersion(rentalInfo.MCVersion)
	if err != nil {
		return 0, enhance.UsingMod{}, nil, &defines.ProtocolError{Message: err.Error()}
	}
	versionCache.SetDefault(req.ServerCode, currentGameInfo.EngineVersion)
	// check version
	if gu.GameInfo.EngineVersion != currentGameInfo.EngineVersion {
		// re-login and get chain with updated engine version
		return requestServerInfo(gu, req)
	}
	return launcherLevel, currentUsingMod, rentalInfo, nil
}

func Login(c *gin.Context) {
	// parse request
	var request AuthRequest
	var helper define.AuthServerHelper
	var enableEncrypt bool

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, AuthResponse{
			SuccessStates: false,
			Message: Message{
				Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
			},
		})
		return
	}

	if database.CheckAuthHelperByToken(request.FBToken, true) {
		helper = database.GetAuthHelperByToken(request.FBToken, true)
	} else {
		splitAns := strings.Split(request.FBToken, "|")
		if len(splitAns) != 2 {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: "Login: 提供的 FB Token 无效",
				},
			})
			return
		}

		eulogistUserUniqueID := splitAns[0]
		helperUniqueID := splitAns[1]

		if !database.CheckUserByUniqueID(eulogistUserUniqueID, true) || !database.CheckAuthHelperByUniqueID(helperUniqueID, true) {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: "Login: 提供的 FB Token 无效",
				},
			})
			return
		}

		accessPass := false
		configs := database.GetAllowServerConfig(request.ServerCode, true)
		for _, value := range configs {
			if value.EulogistUserUniqueID == eulogistUserUniqueID {
				accessPass = true
				break
			}
		}

		if !accessPass {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: fmt.Sprintf(
						"Login: 进入目标租赁服 (%s) 需要租赁服管理人员的授权。如果您没有授权, 可以使用第三方验证服务解决问题",
						request.ServerCode,
					),
				},
			})
			return
		}

		helper = database.GetAuthHelperByUniqueID(helperUniqueID, true)
		enableEncrypt = true
	}

	// dry login
	if request.ServerCode == "::DRY::" && request.ServerPassword == "::DRY::" {
		c.JSON(
			http.StatusOK,
			AuthResponse{
				SuccessStates: true,
				Message:       Message{Information: "ok"},
				FBToken:       request.FBToken,
			},
		)
		return
	}

	// decode to mpay user
	var protocolErr *defines.ProtocolError
	var mu *defines.MpayUser = new(defines.MpayUser)
	var gu *g79.G79User

	if len(request.ProvidedSauthData) == 0 {
		err = json.Unmarshal(helper.MpayUserData, mu)
	} else {
		gu, err = enhance.PEAuthLogin(request.ProvidedSauthData)
	}
	if err != nil {
		c.JSON(http.StatusOK, AuthResponse{
			SuccessStates: false,
			Message: Message{
				Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
			},
		})
		return
	}

	// change engine version by cache
	engineVersion := gameinfo.DefaultEngineVersion
	if value, ok := versionCache.Get(request.ServerCode); ok {
		engineVersion = value.(string)
	}

	// g79 login if this is normal login (but not pe-auth login)
	if gu == nil {
		if gu, protocolErr = g79.Login(engineVersion, mu); protocolErr != nil {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", protocolErr.Error()),
				},
			})
			return
		}
	}

	// request server info
	launcherLevel, currentUsingMod, serverInfo, protocolErr := requestServerInfo(gu, &request)
	if protocolErr != nil {
		c.JSON(http.StatusOK, AuthResponse{
			SuccessStates: false,
			Message: Message{
				Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", protocolErr.Error()),
			},
		})
		return
	}

	// get session
	session := utils.GetSessionByBearer(c)
	if session == nil {
		c.JSON(http.StatusOK, AuthResponse{
			SuccessStates: false,
			Message: Message{
				Information: fmt.Sprintf(
					"Login: 无效的 Auth Bearer (%s)",
					c.Request.Header.Get("Authorization"),
				),
			},
		})
		return
	}

	// save info for anti-cheat callback
	session.Store(session_key_entity_id, gu.EntityID)
	session.Store(session_key_engine_version, gu.GameInfo.EngineVersion)
	session.Store(session_key_patch_version, gu.GameInfo.PatchVersion)

	// response
	resp := AuthResponse{
		SuccessStates:  true,
		Message:        Message{Information: "ok"},
		BotLevel:       launcherLevel,
		BotSkin:        currentUsingMod.AsPhoenixBotSkin(),
		BotComponent:   currentUsingMod.AsPhoenixBotComponent(),
		FBToken:        request.FBToken,
		MasterName:     gu.Username,
		RentalServerIP: serverInfo.IPAddress,
		ChainInfo:      serverInfo.ChainInfo,
	}

	if enableEncrypt {
		jsonBytes, err := json.Marshal(resp)
		if err != nil {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
				},
			})
		}

		encrypted, err := utils.EncryptPKCS1v15(PhoenixLoginKey, jsonBytes)
		if err != nil {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
				},
			})
		}

		c.Data(http.StatusOK, "application/octet-stream", encrypted)
		return
	}

	c.JSON(http.StatusOK, resp)
}
