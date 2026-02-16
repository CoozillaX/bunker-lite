package std_api

import (
	"bunker-lite/enhance"
	"bunker-lite/service/database"
	"bunker-lite/service/define"
	"bunker-lite/service/routers/keys"
	"bunker-lite/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/mpay/android"

	"encoding/hex"
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
	FBToken            string `json:"login_token,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data"`
	ServerCode         string `json:"server_code"`
	ServerPassword     string `json:"server_passcode"`
	ClientPublicKey    string `json:"client_public_key"`
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
	BotSkin      enhance.PhoenixSkinInfo `json:"skin_info"`             // 机器人的皮肤信息
	BotComponent map[string]*int         `json:"outfit_info,omitempty"` // 机器人当前已加载的组件及其附加值

	FBToken        string `json:"token"`           // ...
	MasterName     string `json:"respond_to"`      // 机器人主人的游戏名称
	EnableVitality bool   `json:"enable_vitality"` // 是否启用 Vitality API

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
	isSpecialRequest bool,
	gu *g79.G79User,
	req *AuthRequest,
) (
	launcherLevel int,
	currentUsingMod enhance.UsingMod,
	rentalServerInfo *g79.RentalServerInfo,
	needRelogin bool,
	protocolError *defines.ProtocolError,
) {
	// launcher level
	launcherLevel, _, _, protocolError = enhance.GetLauncherLevel(gu)
	if protocolError != nil {
		return 0, enhance.UsingMod{}, nil, false, protocolError
	}
	// using mod
	currentUsingMod, protocolError = enhance.GetCurrentUsingMod(gu)
	if protocolError != nil {
		return 0, enhance.UsingMod{}, nil, false, protocolError
	}
	// chain info
	rentalInfo, protocolError := gu.ImpactRentalServer(req.ServerCode, req.ServerPassword, req.ClientPublicKey)
	if protocolError != nil {
		return 0, enhance.UsingMod{}, nil, false, protocolError
	}
	// get version
	temp := new(android.AndroidMpayUser)
	if err := temp.UpdateGameInfoByBedrockVersion(strings.TrimSuffix(rentalInfo.MCVersion, "-release")); err != nil {
		return 0, enhance.UsingMod{}, nil, false, &defines.ProtocolError{
			Message: fmt.Sprintf("requestServerInfo: %v", err),
		}
	}
	currentGameInfo := temp.GameInfo
	// cache and check version
	if !isSpecialRequest {
		// do cache
		versionCache.SetDefault(req.ServerCode, currentGameInfo.EngineVersion)
		// need relogin with new engine version
		if gu.GetEngineVersion() != currentGameInfo.EngineVersion {
			return 0, enhance.UsingMod{}, nil, true, nil
		}
	}
	return launcherLevel, currentUsingMod, rentalInfo, false, nil
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

		eulogitUserToken := splitAns[0]
		helperUniqueID := splitAns[1]

		if !database.CheckUserByToken(eulogitUserToken, true) || !database.CheckAuthHelperByUniqueID(helperUniqueID, true) {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: "Login: 提供的 FB Token 无效",
				},
			})
			return
		}

		user := database.GetUserByToken(eulogitUserToken, true)
		if user.UnbanUnixTime >= time.Now().Unix() {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: "Login: 赞颂者用户仍被封禁中",
				},
			})
			return
		}

		accessPass := true // user.CanAccessAnyRentalServer
		configs := database.GetAllowServerConfig(request.ServerCode, true)
		for _, value := range configs {
			if value.EulogistUserUniqueID == user.UserUniqueID {
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

	// ensure engine version
	engineVersion := android.DefaultEngineVersion
	if len(request.ProvidedPeAuthData) == 0 && len(request.ProvidedSaAuthData) == 0 {
		// change version by cache if we use mpay user to login
		value, ok := versionCache.Get(request.ServerCode)
		if ok {
			engineVersion = value.(string)
		}
	}

	// g79 login
	activeSession, err := database.LoadOrRegisterActiveSession(
		helper,
		engineVersion,
		request.ProvidedPeAuthData,
		request.ProvidedSaAuthData,
		true,
		true,
		true,
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, AuthResponse{
			SuccessStates: false,
			Message: Message{
				Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
			},
		})
		return
	}

	// At least loop for two times. If all failed, then
	// result in can not switch to the correct version.
	for range 2 {
		// request server info
		launcherLevel, currentUsingMod, serverInfo, needRelogin, protocolError := requestServerInfo(
			activeSession.SessionType != define.SessionTypeMpayUser,
			activeSession.G79User(),
			&request,
		)
		if protocolError != nil {
			c.JSON(http.StatusOK, AuthResponse{
				SuccessStates: false,
				Message: Message{
					Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", protocolError.Error()),
				},
			})
			return
		}

		// the rental server version is not match, and need relogin
		if needRelogin {
			// change version by cache
			value, ok := versionCache.Get(request.ServerCode)
			if ok {
				engineVersion = value.(string)
			}
			// relogin g79 user
			activeSession, err = database.RegisterActiveSession(
				helper,
				engineVersion,
				"",
				"",
				true,
				true,
			)
			if err != nil {
				c.JSON(http.StatusOK, AuthResponse{
					SuccessStates: false,
					Message: Message{
						Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
					},
				})
				return
			}
			// continue the loop so that we
			// can re-request server info
			continue
		}

		// get tradition session
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
		session.Store(session_key_entity_id, activeSession.G79User().EntityID)
		session.Store(session_key_engine_version, activeSession.G79User().GetEngineVersion())
		session.Store(session_key_patch_version, activeSession.G79User().GetPatchVersion())
		session.Store(session_key_platform, activeSession.G79User().GetSystemName())

		// get helper token if enable encrypt
		helperToken := helper.HelperToken
		if enableEncrypt {
			encryptedToken, err := utils.EncryptPKCS1v15(&keys.TokenEncryptKey.PublicKey, []byte(helper.HelperToken))
			if err != nil {
				c.JSON(http.StatusOK, AuthResponse{
					SuccessStates: false,
					Message: Message{
						Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
					},
				})
				return
			}
			helperToken = hex.EncodeToString(encryptedToken)
		}

		// response
		resp := AuthResponse{
			SuccessStates:  true,
			Message:        Message{Information: "ok"},
			BotLevel:       launcherLevel,
			BotSkin:        currentUsingMod.AsPhoenixBotSkin(),
			BotComponent:   currentUsingMod.AsPhoenixBotComponent(),
			FBToken:        helperToken,
			EnableVitality: helper.EnableVitality,
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
				return
			}

			encrypted, err := utils.EncryptPKCS1v15(keys.PhoenixLoginKey, jsonBytes)
			if err != nil {
				c.JSON(http.StatusOK, AuthResponse{
					SuccessStates: false,
					Message: Message{
						Information: fmt.Sprintf("Login: 登录到租赁服时出现问题, 原因是 %v", err),
					},
				})
				return
			}

			c.Data(http.StatusOK, "application/octet-stream", encrypted)
			return
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		SuccessStates: false,
		Message: Message{
			Information: "Login: 目标版本的租赁服不支持, 请等待地堡适配",
		},
	})
}
