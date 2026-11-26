package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) getP2PConfigs(GAorB []byte) (*P2PConfig, error) {
	dhConfigRaw, err := ctx.app.MessagesGetDhConfig(0, 256)
	if err != nil {
		return nil, err
	}
	dhConfig, ok := dhConfigRaw.(*tg.MessagesDhConfigObj)
	if !ok {
		return nil, fmt.Errorf("unexpected dh config type %T", dhConfigRaw)
	}
	return &P2PConfig{
		DhConfig:   dhConfig,
		IsOutgoing: GAorB == nil,
		GAorB:      GAorB,
		WaitData:   make(chan error, 1),
	}, nil
}
