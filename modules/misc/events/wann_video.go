package events

import (
	"strings"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

func init() {
	core.On(wannVideoHandler)
}

func wannVideoHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	msg := strings.ToLower(m.Content)

	if strings.Contains(msg, "wann") && strings.Contains(msg, "video") {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "https://youtu.be/4btyXfex_8w", m.Reference())
	}
}
