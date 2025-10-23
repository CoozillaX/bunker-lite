package main

// ------------------------- Register G79 User Transaction -------------------------

// RegisterActiveGuRequest ..
type RegisterActiveGuRequest struct {
	Token              string `json:"token,omitempty"`
	OverrideSession    bool   `json:"override_session,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data,omitempty"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data,omitempty"`
}

// RegisterActiveGuResponse ..
type RegisterActiveGuResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionID         string `json:"session_id"`
	SessionType       uint8  `json:"session_type"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// ------------------------- Request Session Info -------------------------

const (
	ResponseTypeFindSession uint8 = iota
	ResponseTypeNoSession
)

// SessionInfoRequest ..
type SessionInfoRequest struct {
	Token string `json:"token,omitempty"`
}

// SessionInfoResponse ..
type SessionInfoResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	ResponseType      uint8  `json:"response_type"`
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

// ------------------------- Request Daily Growth -------------------------

// DailyGrowthRequest ..
type DailyGrowthRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// DailyGrowthResponse ..
type DailyGrowthResponse struct {
	ErrorInfo      string `json:"error_info"`
	Success        bool   `json:"success"`
	XpFromOnline   int    `json:"xp_from_online"`
	XpFromRecharge int    `json:"xp_from_recharge"`
}

// ------------------------- Get Currency Online -------------------------

// CurrencyOnlineRequest ..
type CurrencyOnlineRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// CurrencyOnlineResponse ..
type CurrencyOnlineResponse struct {
	ErrorInfo        string `json:"error_info"`
	Success          bool   `json:"success"`
	RestCurrencyTime int    `json:"rest_currency_time"`
	FormatDateString string `json:"format_date_string"`
}

// ------------------------- Keep G79 User Alive -------------------------

const (
	KeepGuAliveErrorMeetError uint8 = iota
	KeepGuAliveErrorLifeLimit
)

// KeepGuAliveRequest ..
type KeepGuAliveRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// KeepGuAliveResponse ..
type KeepGuAliveResponse struct {
	ErrorType         uint8  `json:"error_type"`
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// ------------------------- End -------------------------
