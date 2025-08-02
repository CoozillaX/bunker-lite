package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type RentalServerConfig struct {
	ServerNumber   string
	ServerPassCode string
}

func (r *RentalServerConfig) Marshal(io protocol.IO) {
	io.String(&r.ServerNumber)
	io.String(&r.ServerPassCode)
}
