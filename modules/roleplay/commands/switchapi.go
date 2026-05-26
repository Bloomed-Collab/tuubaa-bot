package commands

import (
	"github.com/S42yt/tuubaa-bot/modules/roleplay/api"
	"github.com/bwmarrin/discordgo"
)

func SwitchAPIHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data := i.ApplicationCommandData()
		if len(data.Options) == 0 {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Please provide an API choice."},
			})
		}
		choice := data.Options[0].StringValue()
		var msg string
		if !api.IsBastiAvailable() && (choice != "Otaku") {
			api.SetAPItype(api.OTAKU)
			msg = "Basti API is not available, defaulting to OtakuGIFs API."
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: msg},
			})
		}
		switch choice {
		case "Otaku":
			api.SetAPItype(api.OTAKU)
			msg = "Switched to OtakuGIFs API."
		case "Basti":
			api.SetAPItype(api.BASTI)
			msg = "Switched to Bastiwood API."
		case "Both":
			api.SetAPItype(api.BOTH)
			msg = "Switched to Both APIs (random)."
		default:
			msg = "Unknown API choice."
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: msg},
		})
	}
}
