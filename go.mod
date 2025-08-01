module bunker-lite

go 1.24.0

require (
	bunker-core v0.0.0
	github.com/google/uuid v1.6.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
)

require (
	github.com/go-gl/mathgl v1.1.0 // indirect
	golang.org/x/image v0.21.0 // indirect
)

require (
	github.com/database64128/chacha8-go v0.0.0-20250204235950-5c6f473ea976 // indirect
	github.com/sandertv/gophertunnel v1.48.1
	golang.org/x/sys v0.30.0 // indirect
)

replace bunker-core => ../bunker-core
