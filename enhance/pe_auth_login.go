package enhance

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// PEAuthLogin make user login by peAuthStringData.
// This option will have no matters to the data that
// saved in the eulogist database.
//
// Provided peAuthStringData must be json string or
// hex string. If is a hex string, it means that
// peAuthStringData is created by user that they use
// capture software to get pe-authentication http packet.
//
// Note that gu.MpayUser is not very completely, so
// that the returned g79 user only can be used once.
func PEAuthLogin(peAuthStringData string) (gu *g79.G79User, err error) {
	// 0. Prepare
	var saAuthData map[string]any
	var saDataInCookie map[string]any
	var validData string

	// 1. Unmarshal json string
	err = json.Unmarshal([]byte(peAuthStringData), &saAuthData)
	if err != nil {
		// err is not nil is maybe the user directly provided captured
		// pe-authentication packet.
		// Therefore, we need decrypt it and do some prefix.
		encryptedBytes, err := hex.DecodeString(peAuthStringData)
		if err != nil {
			return nil, fmt.Errorf("PEAuthLogin: %v", err)
		}
		// decrypt the packet
		originData, err := g79.HttpDecrypt(encryptedBytes)
		if err != nil {
			return nil, fmt.Errorf("PEAuthLogin: %v", err)
		}
		// do prefix
		validData = string(originData)
		for {
			if len(validData) == 0 {
				break
			}
			if validData[len(validData)-1] == '}' {
				break
			}
			validData = validData[:len(validData)-1]
		}
		// unmarshal again
		err = json.Unmarshal([]byte(validData), &saAuthData)
		if err != nil {
			return nil, fmt.Errorf("PEAuthLogin: %v", err)
		}
	}

	// 2. Get and basic data
	engineVersion, exist1 := saAuthData["engine_version"].(string)
	patchVersion, exist2 := saAuthData["patch_version"].(string)
	saDataString, _ := saAuthData["sa_data"].(string)
	_ = json.Unmarshal([]byte(saDataString), &saDataInCookie)
	cpuDigit, exist3 := saDataInCookie["cpu_digit"].(string)
	osName, exist4 := saDataInCookie["os_name"].(string)
	if !exist1 || !exist2 || !exist3 || !exist4 {
		return nil, fmt.Errorf("PEAuthLogin: Wrong PE Auth data string %#v", peAuthStringData)
	}

	// 3. Sync part of basic data
	defaultBaseInfo, err := gameinfo.GetInfoByEngineVersion(gameinfo.DefaultEngineVersion)
	if err != nil {
		return nil, fmt.Errorf("PEAuthLogin: %v", err)
	}
	copiedBaseInfo := *defaultBaseInfo
	defaultBaseInfo = &copiedBaseInfo
	defaultBaseInfo.EngineVersion = engineVersion
	defaultBaseInfo.PatchVersion = patchVersion

	// 4. Get auth message but not included seed
	authMessage, exist := saAuthData["message"].(string)
	if !exist || len(authMessage) < 36 {
		return nil, fmt.Errorf("PEAuthLogin: Wrong auth message %#v was found", authMessage)
	}
	authMessage = authMessage[:len(authMessage)-36]

	// 5. Generate new seed and put it to auth message
	seed := uuid.NewString()
	authMessage += seed

	// 6. Prefix saAuthData and marshal
	saAuthData["seed"] = seed
	saAuthData["message"] = authMessage
	saAuthData["sign"] = g79.CalculateAuthenticationSign(authMessage, defaultBaseInfo.AuthSignKey, defaultBaseInfo.AuthSignCycle)
	reqBody, _ := json.Marshal(saAuthData)

	// 7. Do request
	gu = &g79.G79User{
		MpayUser: defines.MpayUser{
			MpayDevice: defines.MpayDevice{
				CPUDigit:   cpuDigit,
				SystemName: osName,
			},
		},
		GameInfo: defaultBaseInfo,
	}
	reader, protocolErr := gu.CreateHttpClient().
		SetMethod(http.MethodPost).
		SetUrl(gameinfo.G79ServerList.CoreServerUrl + "/pe-authentication").
		SetRawBody(reqBody).
		SetEncryptSuffix(0xc).
		Do()
	if protocolErr != nil {
		return nil, protocolErr
	}

	// 8. Parse response
	var query struct {
		Entity *g79.G79User `json:"entity"`
	}
	if err := json.NewDecoder(reader).Decode(&query); err != nil {
		return nil, &defines.ProtocolError{Message: err.Error()}
	}
	gu = query.Entity
	gu.MpayUser.MpayDevice.CPUDigit = cpuDigit
	gu.MpayUser.MpayDevice.SystemName = osName
	gu.GameInfo = defaultBaseInfo

	// 9. Get user name
	{
		reqBody, _ := json.Marshal(map[string]string{
			"entity_id": gu.EntityID,
		})
		reader, protocolErr := gu.CreateHttpClient().
			SetMethod(http.MethodPost).
			SetUrl(gameinfo.G79ServerList.CoreServerUrl + "/pe-user-detail/get").
			SetRawBody(reqBody).
			SetTokenMode(g79.TOKEN_MODE_NORMAL).
			Do()
		if protocolErr != nil {
			return nil, protocolErr
		}
		var respEntity struct {
			Entity *struct {
				Name string `json:"name"`
			} `json:"entity"`
		}
		if err := json.NewDecoder(reader).Decode(&respEntity); err != nil {
			return nil, &defines.ProtocolError{Message: err.Error()}
		}
		gu.Username = respEntity.Entity.Name
	}

	return gu, nil
}
