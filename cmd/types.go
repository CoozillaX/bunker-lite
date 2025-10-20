package main

// ------------------------- Register G79 User Transaction -------------------------

// RegisterActiveGuRequest ..
type RegisterActiveGuRequest struct {
	Token              string `json:"token,omitempty"`
	OverrideSession    bool   `json:"override_session,omitempty"`
	EngineVersion      string `json:"engine_version,omitempty"`
	ProvidedPeAuthData string `json:"provided_pe_auth_data,omitempty"`
	ProvidedSaAuthData string `json:"provided_sa_auth_data,omitempty"`
}

// RegisterActiveGuResponse ..
type RegisterActiveGuResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionID         string `json:"session_id"`
	SessionExpireTime int64  `json:"session_expire_time"`
}

// ------------------------- Keep G79 User Alive -------------------------

// KeepGuAliveRequest ..
type KeepGuAliveRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// KeepGuAliveResponse ..
type KeepGuAliveResponse struct {
	ErrorInfo         string `json:"error_info"`
	Success           bool   `json:"success"`
	SessionExpireTime int64  `json:"session_expire_time"`
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

// ------------------------- Request Daily Growth -------------------------

// DailyGrowthRequest ..
type DailyGrowthRequest struct {
	Token     string `json:"token,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// DailyGrowthResponse ..
type DailyGrowthResponse struct {
	ErrorInfo    string `json:"error_info"`
	Success      bool   `json:"success"`
	XpFromOnline int    `json:"xp_from_online"`
	XpFromCharge int    `json:"xp_from_charge"`
}

// ------------------------- End -------------------------
