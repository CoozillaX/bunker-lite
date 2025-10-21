package define

const (
	StdAuthServerPhoenixAPI = "https://yorha.eulogist-api.icu"
	StdAuthServerAddress    = "https://yorha.eulogist-api.icu/eulogist_api"
	UserPasswordSalt        = "YoRHa"
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
