package define

const (
	StdAuthServerAddress = "http://127.0.0.1:8080"
	UserPasswordSlat     = "YoRHa"
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
