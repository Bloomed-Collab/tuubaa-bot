package LLM

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// messageHandler fires on every message. Queues a response if the bot
// was mentioned or if the message is a reply to the bot.
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !loaded {
		return
	}
	if m.Author == nil || m.Author.Bot {
		return
	}

	botID := s.State.User.ID

	mentioned := false
	for _, u := range m.Mentions {
		if u.ID == botID {
			mentioned = true
			break
		}
	}

	isReply := m.ReferencedMessage != nil &&
		m.ReferencedMessage.Author != nil &&
		m.ReferencedMessage.Author.ID == botID

	if !mentioned && !isReply {
		return
	}

	content := strings.TrimSpace(strings.ReplaceAll(m.Content, "<@"+botID+">", ""))
	if content == "" {
		return
	}

	select {
	case queue <- queueItem{s: s, channelID: m.ChannelID, messageID: m.ID, message: content}:
	default:
		s.ChannelMessageSend(m.ChannelID, "I'm busy, please try again in a moment.")
	}
}
