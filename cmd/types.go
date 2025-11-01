package main

// ------------------------- Register Active Session -------------------------

// RegisterActiveSessionRequest ..
type RegisterActiveSessionRequest struct {
	Token              string `json:"token,omitempty"`
	OverrideSession    bool   `json:"override_session,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data,omitempty"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data,omitempty"`
}

// RegisterActiveSessionResponse ..
type RegisterActiveSessionResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionID         string `json:"session_id"`
	SessionType       uint8  `json:"session_type"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// ------------------------- Clean Up Session -------------------------

// CleanUpSessionRequest ..
type CleanUpSessionRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// CleanUpSessionResponse ..
type CleanUpSessionResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// ------------------------- Keep Session Alive -------------------------

const (
	KeepSessionAliveErrorMeetError uint8 = iota
	KeepSessionAliveErrorLifeLimit
)

// KeepSessionAliveRequest ..
type KeepSessionAliveRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// KeepSessionAliveResponse ..
type KeepSessionAliveResponse struct {
	ErrorType         uint8  `json:"error_type"`
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// ------------------------- Request Vitality Debug -------------------------

const (
	RequestTypeGetCurrencyOnline uint8 = iota
	RequestTypeGetDailyGrowth
)

// VitalityDebugRequest ..
type VitalityDebugRequest struct {
	Token       string `json:"token,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	RequestType uint8  `json:"request_type,omitempty"`
}

// VitalityDebugResponse ..
type VitalityDebugResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`

	RestCurrencyTime int    `json:"rest_currency_time"`
	FormatDateString string `json:"format_date_string"`

	XpFromOnline   int `json:"xp_from_online"`
	XpFromRecharge int `json:"xp_from_recharge"`
}

// ------------------------- End -------------------------
