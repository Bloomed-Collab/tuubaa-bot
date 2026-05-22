package commands

import (
	"github.com/S42yt/tuubaa-bot/modules/chatgpt/handler"
	"github.com/bwmarrin/discordgo"
)

// ChangeAIHandler returns the handler for /changeai.
func ChangeAIHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options
		if len(opts) == 0 {
			return respondText(s, i, "Missing backend option.")
		}
		switch opts[0].StringValue() {
		case "openai":
			handler.ActiveBackend = handler.BackendOpenAI
			return respondText(s, i, "✅ Switched to **ChatGPT (OpenAI)** backend.")
		case "basti":
			handler.ActiveBackend = handler.BackendBasti
			if !handler.IsBastiLoaded() {
				return respondText(s, i, "✅ Switched to **Basti API** backend. The model is not loaded yet — use `/loadai` first.")
			}
			return respondText(s, i, "✅ Switched to **Basti API** backend.")
		case "disabled":
			handler.ActiveBackend = handler.BackendDisabled
			return respondText(s, i, "🔴 AI is now **disabled** globally.")
		default:
			return respondText(s, i, "Unknown backend.")
		}
	}
}

// LoadAIHandler returns the handler for /loadai.
func LoadAIHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if err := handler.LoadBastiLLM(); err != nil {
			return respondText(s, i, "❌ Failed to load Basti model: "+err.Error())
		}
		return respondText(s, i, "✅ Basti model loaded.")
	}
}

// UnloadAIHandler returns the handler for /unloadai.
func UnloadAIHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if err := handler.UnloadBastiLLM(); err != nil {
			return respondText(s, i, "❌ Failed to unload Basti model: "+err.Error())
		}
		return respondText(s, i, "✅ Basti model unloaded.")
	}
}

// SetPromptHandler returns the handler for /setprompt.
func SetPromptHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options
		if len(opts) == 0 {
			return respondText(s, i, "Missing prompt text.")
		}
		handler.SetBastiPrompt(opts[0].StringValue())
		return respondText(s, i, "✅ Basti system prompt updated.")
	}
}

func respondText(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
