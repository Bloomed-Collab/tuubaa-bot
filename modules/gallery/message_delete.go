package gallery

import (
	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func messageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	if m.GuildID == "" {
		return
	}

	artChannels, err := cfg.GetArtChannels(m.GuildID)
	if err != nil || !containsStr(artChannels, m.ChannelID) {
		return
	}

	threadID, postID, err := deletePost(m.GuildID, m.ChannelID, m.ID)
	if err != nil {
		logger.Warn("gallery: messageDelete deletePost: %v", err)
		return
	}
	if threadID == "" {
		return
	}

	if err := s.ChannelMessageDelete(threadID, postID); err != nil {
		logger.Warn("gallery: messageDelete delete gallery post %s: %v", postID, err)
	}
}
