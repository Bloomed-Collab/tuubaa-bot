package gallery

import "github.com/S42yt/tuubaa-bot/core"

func init() {
	core.On(reactionAddHandler)
	core.On(reactionRemoveHandler)
	registerGalleryCommand()
}
