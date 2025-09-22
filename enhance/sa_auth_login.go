package enhance

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-core/utils"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// SaAuthData ..
type SaAuthData struct {
	MpayID    string `json:"deviceid"`
	Uid       string `json:"sdkuid"`
	MpayToken string `json:"sessionid"`
	Udid      string `json:"udid"`
}

// SaAuthLogin ..
func SaAuthLogin(saAuthJsonData string) (gu *g79.G79User, err error) {
	// 0. Prepare
	var saAuthData SaAuthData

	// 1. Unmarshal json string
	err = json.Unmarshal([]byte(saAuthJsonData), &saAuthData)
	if err != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", err)
	}

	// 2. Create mu and sync data
	mu := new(defines.MpayUser)
	mu.Uid = saAuthData.Uid
	mu.MpayToken = saAuthData.MpayToken
	mu.MpayDevice.Udid = saAuthData.Udid
	mu.MpayDevice.MpayID = saAuthData.MpayID

	// 3. Generate device data
	mu.MpayDevice.UniqueId = uuid.NewString() + utils.GenerateTimestampWithLength(13)
	mu.MpayDevice.UrsUdid = utils.GenerateRandomString(39)
	mu.MpayDevice.Mac = utils.GenerateRandomString(32)
	mu.MpayDevice.MACAddr = "02:00:00:00:00:00"
	mu.MpayDevice.InitUrsDevice = "0"
	mu.MpayDevice.DeviceType = "mobile"
	mu.MpayDevice.Brand = "Huawei"
	mu.MpayDevice.DeviceName = strings.ToUpper("Huawei" + "_" + utils.GenerateRandomString(6))
	mu.MpayDevice.DeviceModel = mu.MpayDevice.DeviceName
	mu.MpayDevice.SystemName = "Android"
	mu.MpayDevice.SystemVersion = strconv.Itoa(7 + rand.N(5))
	mu.MpayDevice.CoreNum = "\\b"
	mu.MpayDevice.CPUDigit = "64"
	mu.MpayDevice.CPUHz = fmt.Sprintf("%d", (rand.N(2000)+6000)*100)
	mu.MpayDevice.CPUName = "Snapdragon 888"
	mu.MpayDevice.Resolution = "1920*1080"
	mu.MpayDevice.DeviceWidth = strings.Split(mu.MpayDevice.Resolution, "*")[0]
	mu.MpayDevice.DeviceHeight = strings.Split(mu.MpayDevice.Resolution, "*")[1]
	mu.MpayDevice.Disk = ""
	mu.MpayDevice.Emulator = 0
	mu.MpayDevice.Network = "CHANNEL_UNKNOW"
	mu.MpayDevice.RAM = fmt.Sprintf("%d", rand.N(int64(3e9))+1e9)
	mu.MpayDevice.ROM = fmt.Sprintf("%d", rand.N(int64(115e9))+5e9)
	mu.MpayDevice.Root = false

	// Get base info
	defaultBaseInfo, err := gameinfo.GetInfoByEngineVersion(gameinfo.DefaultEngineVersion)
	if err != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", err)
	}

	// g79 login
	gu, protocolErr := g79.Login(defaultBaseInfo.EngineVersion, mu)
	if protocolErr != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", protocolErr.Error())
	}

	return gu, nil
}
