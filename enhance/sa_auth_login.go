package enhance

import (
	"bunker-core/protocol/g79"
	"bunker-core/protocol/mpay/android"
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
func SaAuthLogin(engineVersion string, saAuthJsonData string) (g79User *g79.G79User, err error) {
	// 0. Prepare
	var saAuthData SaAuthData
	var mpayUser android.AndroidMpayUser

	// 1. Unmarshal json string
	err = json.Unmarshal([]byte(saAuthJsonData), &saAuthData)
	if err != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", err)
	}

	// 2. Init mpay user
	protocolErr := mpayUser.Initialise()
	if protocolErr != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", protocolErr.Error())
	}

	// 3. Sync data
	mpayUser.Uid = saAuthData.Uid
	mpayUser.MpayToken = saAuthData.MpayToken
	mpayUser.AndroidMpayDevice.Udid = saAuthData.Udid
	mpayUser.AndroidMpayDevice.MpayID = saAuthData.MpayID

	// 4. g79 login
	g79User, protocolError := g79.Login(&mpayUser)
	if protocolError != nil {
		return nil, fmt.Errorf("SaAuthLogin: %v", protocolError.Error())
	}
	return g79User, nil
}
