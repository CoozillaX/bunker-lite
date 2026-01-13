package utils

import (
	"bunker-core/protocol/mpay/android"
	"encoding/base64"
	"encoding/json"
)

func EncodeFBToken(mu *android.AndroidMpayUser) string {
	raw, _ := json.Marshal(mu)
	return base64.StdEncoding.EncodeToString(raw)
}

func DecodeFBToken(token string) (*android.AndroidMpayUser, error) {
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var mu android.AndroidMpayUser
	if err := json.Unmarshal(raw, &mu); err != nil {
		return nil, err
	}
	return &mu, nil
}
