package gallery

import (
	"fmt"
	"strings"

	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func reactionAddHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
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
	if r.UserID != msg.Author.ID {
		return
	}

	var imageURLs []string
	for _, a := range msg.Attachments {
		if strings.HasPrefix(a.ContentType, "image/") || isImageURL(a.URL) {
			imageURLs = append(imageURLs, a.URL)
		}
	}
	if len(imageURLs) == 0 {
		return
	}

	existing, err := getPost(r.GuildID, r.ChannelID, r.MessageID)
	if err != nil {
		logger.Warn("gallery: getPost: %v", err)
		return
	}
	if existing != nil {
		return
	}

	forumID, err := cfg.GetChannel(r.GuildID, "gallery_forum")
	if err != nil || forumID == "" {
		return
	}

	displayName := resolveDisplayName(s, r.GuildID, r.UserID)

	threadID, err := getThread(r.GuildID, r.UserID)
	if err != nil {
		logger.Warn("gallery: getThread: %v", err)
		return
	}

	messageLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", r.GuildID, r.ChannelID, r.MessageID)

	starCount := 1
	for _, reaction := range msg.Reactions {
		if reaction.Emoji.Name == "⭐" {
			starCount = reaction.Count
			break
		}
	}

	starDisplay := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("⭐ %d", starCount)).Build()
	container := buildGalleryContainer(msg.Content, imageURLs, msg.Timestamp, messageLink)
	send := &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{starDisplay, container},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}

	var postID string

	if threadID == "" {
		thread, err := s.ForumThreadStart(forumID, displayName, 10080, "-# 🖼️ "+displayName)
		if err != nil {
			logger.Warn("gallery: create thread for %s: %v", r.UserID, err)
			return
		}
		threadID = thread.ID
		if err := saveThread(r.GuildID, r.UserID, threadID); err != nil {
			logger.Warn("gallery: saveThread: %v", err)
		}
		locked := true
		if _, err := s.ChannelEditComplex(threadID, &discordgo.ChannelEdit{Locked: &locked}); err != nil {
			logger.Warn("gallery: lock thread %s: %v", threadID, err)
		}
	}

	posted, err := s.ChannelMessageSendComplex(threadID, send)
	if err != nil {
		logger.Warn("gallery: post to thread %s: %v", threadID, err)
		return
	}
	postID = posted.ID

	if err := savePost(r.GuildID, r.ChannelID, r.MessageID, threadID, postID); err != nil {
		logger.Warn("gallery: savePost: %v", err)
	}
}

func resolveDisplayName(s *discordgo.Session, guildID, userID string) string {
	member, err := s.GuildMember(guildID, userID)
	if err != nil || member == nil {
		return userID
	}
	if member.Nick != "" {
		return member.Nick
	}
	if member.User != nil {
		if member.User.GlobalName != "" {
			return member.User.GlobalName
		}
		return member.User.Username
	}
	return userID
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func isImageURL(url string) bool {
	l := strings.ToLower(url)
	return strings.HasSuffix(l, ".jpg") || strings.HasSuffix(l, ".jpeg") ||
		strings.HasSuffix(l, ".png") || strings.HasSuffix(l, ".gif") ||
		strings.HasSuffix(l, ".webp")
}
