package events

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/chatgpt/commands"
	"github.com/S42yt/tuubaa-bot/modules/chatgpt/handler"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func init() {
	ulog.Debug("[AI EVENTS] Registering message event handler")
	core.On(messageCreateHandler)
	ulog.Debug("[AI EVENTS] Message event handler registered")
}

func messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	ulog.Debug("[AI] Message event triggered from %s: %s", m.Author.Username, m.Content)

	if m.Author.Bot {
		ulog.Debug("[AI] Ignoring bot message from %s", m.Author.Username)
		return
	}

	if m.GuildID == "" {
		ulog.Debug("[AI] Message not in a guild")
		return
	}

	if s == nil || s.State == nil || s.State.User == nil {
		ulog.Error("[AI] Session state is nil")
		return
	}
	ulog.Debug("[AI] Bot ID: %s", s.State.User.ID)

	if !isBotMentioned(s, m) {
		ulog.Debug("[AI] Bot not mentioned in message. Mentions count: %d", len(m.Mentions))
		return
	}
	ulog.Debug("[AI] Bot mentioned! Guild: %s, Channel: %s, Author: %s", m.GuildID, m.ChannelID, m.Author.Username)

	enabled, err := commands.IsAIEnabled(m.GuildID)
	if err != nil {
		ulog.Warn("[AI] Error checking AI status: %v", err)
		return
	}
	ulog.Debug("[AI] AI Enabled for guild: %v", enabled)
	if !enabled {
		ulog.Debug("[AI] AI is disabled for this guild")
		return
	}

	prompt := extractPrompt(s.State.User.ID, m.Content)
	ulog.Debug("[AI] Extracted prompt: %s", prompt)

	cleanPrompt, err := handler.ValidateAndCleanPrompt(prompt)
	if err != nil {
		ulog.Warn("[AI] Prompt validation error: %v", err)
		sendMessage(s, m.ChannelID, fmt.Sprintf("❌ %s", err.Error()))
		return
	}
	ulog.Debug("[AI] Clean prompt (%d chars): %s", len(cleanPrompt), cleanPrompt)

	handler.AddMessageToCache(m.ChannelID, "user", cleanPrompt)
	ulog.Debug("[AI] Added user message to cache")

	_ = s.ChannelTyping(m.ChannelID)

	ulog.Debug("[AI] Calling OpenAI API with conversation history...")
	response, err := handler.GetAIResponseWithHistory(cleanPrompt, m.ChannelID)
	if err != nil {
		ulog.Error("[AI] API error: %v", err)
		sendMessage(s, m.ChannelID, fmt.Sprintf("❌ Fehler bei der Verarbeitung: %v", err))
		return
	}
	ulog.Debug("[AI] Got response from API: %s", response)

	ulog.Debug("[AI] Response ready to send (not cached)")

	ulog.Debug("[AI] Sending formatted response to channel")
	sendMessage(s, m.ChannelID, response)
	ulog.Debug("[AI] Response sent successfully")
}

func isBotMentioned(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if len(m.Mentions) == 0 {
		return false
	}

	botUserID := s.State.User.ID
	for _, mention := range m.Mentions {
		if mention.ID == botUserID {
			return true
		}
	}
	return false
}

func extractPrompt(botID string, content string) string {
	content = strings.ReplaceAll(content, fmt.Sprintf("<@%s>", botID), "")
	content = strings.ReplaceAll(content, fmt.Sprintf("<@!%s>", botID), "")

	re := regexp.MustCompile(`<@!?\d+>`)
	content = re.ReplaceAllString(content, "")

	return strings.TrimSpace(content)
}

func sendMessage(s *discordgo.Session, channelID string, content string) {
	ulog.Debug("[AI] sendMessage called with content length: %d", len(content))
	if len(content) > 2000 {
		ulog.Warn("[AI] Message truncated from %d to 2000 chars", len(content))
		content = content[:1997] + "..."
	}

	ulog.Debug("[AI] Sending message to channel %s", channelID)
	_, err := s.ChannelMessageSend(channelID, content)
	if err != nil {
		ulog.Error("[AI] Error sending message: %v", err)
		return
	}
	ulog.Debug("[AI] Message sent successfully")
}
