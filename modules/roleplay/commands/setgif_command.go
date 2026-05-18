package commands

import (
	"strings"

	"github.com/S42yt/tuubaa-bot/modules/roleplay/api"
	"github.com/bwmarrin/discordgo"
)

func SetGifHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options
		if len(opts) < 2 {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Usage: /setgif reaction:<name> url:<gif-url>",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		reaction := strings.TrimSpace(opts[0].StringValue())
		gifURL := strings.TrimSpace(opts[1].StringValue())
		if reaction == "" || gifURL == "" {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Reaction and URL must not be empty.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		if err := api.SetGifURL(reaction, gifURL); err != nil {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to save GIF: " + err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Saved GIF for reaction `" + reaction + "`.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
