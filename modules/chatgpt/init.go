package chatgpt

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/chatgpt/commands"
	_ "github.com/S42yt/tuubaa-bot/modules/chatgpt/events"
)

func init() {
	toggleCmd := &core.Command{
		Name:        "ai",
		Description: "Toggle AI responses on/off",
		Options:     nil,
		AllowAdmin:  true,
		Handler:     commands.AIToggleHandler(),
	}

	_ = core.Register(toggleCmd)
}
