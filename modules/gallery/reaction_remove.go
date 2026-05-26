package gallery

import (
	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func reactionRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.Emoji.Name != "⭐" || r.GuildID == "" {
		return
	}

	artChannels, err := cfg.GetArtChannels(r.GuildID)
	if err != nil || !containsStr(artChannels, r.ChannelID) {
		return
	}

	msg, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil || msg.Author == nil {
		return
	}
	existing, err := getPost(r.GuildID, r.ChannelID, r.MessageID)
	if err != nil {
		logger.Warn("gallery: getPost: %v", err)
		return
	}

	// Non-author removing their star: just update the count
	if r.UserID != msg.Author.ID {
		if existing != nil {
			updateGalleryStarCount(s, msg, existing)
		}
		return
	}

	// Author removing their star: delete the gallery post
	threadID, postID, err := deletePost(r.GuildID, r.ChannelID, r.MessageID)
	if err != nil {
		logger.Warn("gallery: deletePost: %v", err)
		return
	}
	if threadID == "" {
		return
	}

	if err := s.ChannelMessageDelete(threadID, postID); err != nil {
		logger.Warn("gallery: delete message %s from thread %s: %v", postID, threadID, err)
	}
}
