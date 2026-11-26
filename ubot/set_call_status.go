package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"

	"KanhaMusic/ntgcalls"
)

func (ctx *Context) setCallStatus(
	call tg.InputGroupCall,
	state ntgcalls.MediaState,
) error {
	if ctx.self == nil {
		return fmt.Errorf("client self user not initialized")
	}
	_, err := ctx.app.PhoneEditGroupCallParticipant(
		&tg.PhoneEditGroupCallParticipantParams{
			Call: call,
			Participant: &tg.InputPeerUser{
				UserID:     ctx.self.ID,
				AccessHash: ctx.self.AccessHash,
			},
			Muted:              state.Muted,
			VideoPaused:        state.VideoPaused,
			VideoStopped:       state.VideoStopped,
			PresentationPaused: state.PresentationPaused,
		},
	)
	return err
}
