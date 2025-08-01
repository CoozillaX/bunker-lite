package define

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// GameSavesKey ..
type GameSavesKey struct {
	EulogistUserUniqueID string
	RentalServerNumber   string
	GameSavesAESKeyBytes []byte
}

func (g *GameSavesKey) MarshalKey(io protocol.IO) {
	io.String(&g.EulogistUserUniqueID)
	io.String(&g.RentalServerNumber)
}

func (g *GameSavesKey) MarshalData(io protocol.IO) {
	io.ByteSlice(&g.GameSavesAESKeyBytes)
}
