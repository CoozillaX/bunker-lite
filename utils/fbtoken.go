package utils

import (
	"bunker-core/protocol/defines"
	"encoding/base64"
	"encoding/json"
)

func EncodeFBToken(mu *defines.MpayUser) string {
	raw, _ := json.Marshal(mu)
	return base64.StdEncoding.EncodeToString(raw)
}

func DecodeFBToken(token string) (*defines.MpayUser, error) {
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var mu defines.MpayUser
	if err := json.Unmarshal(raw, &mu); err != nil {
		return nil, err
	}
	return &mu, nil
}
