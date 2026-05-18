package LLM

import "github.com/S42yt/tuubaa-bot/core"

func init() {
	startWorker()
	core.On(messageHandler)
	registerCommands()
}
