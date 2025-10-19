package enhance

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/mpay"
	"encoding/json"
	"fmt"
)

// SaAuthData ..
type SaAuthData struct {
	MpayID    string `json:"deviceid"`
	Uid       string `json:"sdkuid"`
	MpayToken string `json:"sessionid"`
	Udid      string `json:"udid"`
}

// SaAuthLogin ..
func SaAuthLogin(engineVersion string, saAuthJsonData string) (gu *g79.G79User, err error) {
	// 0. Prepare
	var saAuthData SaAuthData

	// 1. Unmarshal json string
	err = json.Unmarshal([]byte(saAuthJsonData), &saAuthData)
	if err != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", err)
	}

	// 2. Create mu and sync data
	mu := mpay.GenerateMpayUser()
	mu.Uid = saAuthData.Uid
	mu.MpayToken = saAuthData.MpayToken
	mu.MpayDevice.Udid = saAuthData.Udid
	mu.MpayDevice.MpayID = saAuthData.MpayID

	// 3. g79 login
	gu, protocolError := g79.Login(engineVersion, mu)
	if protocolError != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", protocolError.Error())
	}

	return gu, nil
}
