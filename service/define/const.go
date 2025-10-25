package define

const (
	AddressPhoenixAPI  = "https://yorha.eulogist-api.icu"
	AddressEulogistAPI = "https://yorha.eulogist-api.icu/eulogist_api"
	AddressVitalityAPI = "https://yorha.eulogist-api.icu/vitality_api"
	UserPasswordSalt   = "YoRHa"
)

const (
	AuthServerAccountTypeStd uint8 = iota
	AuthServerAccountTypeCustom
)

const (
	UserPermissionSystem = iota
	UserPermissionAdmin
	UserPermissionAdvance
	UserPermissionNormal
	UserPermissionNone
	UserPermissionDefault = UserPermissionNormal
)
