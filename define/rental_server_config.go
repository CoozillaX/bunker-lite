package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type RentalServerConfig struct {
	ServerNumber       string
	ServerPassCode     string
	CanAccessWithoutOP bool
}

func (r *RentalServerConfig) Marshal(io protocol.IO) {
	io.String(&r.ServerNumber)
	io.String(&r.ServerPassCode)
	io.Bool(&r.CanAccessWithoutOP)
}
