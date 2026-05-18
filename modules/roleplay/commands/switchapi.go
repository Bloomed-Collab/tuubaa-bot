package commands

import (
	"github.com/S42yt/tuubaa-bot/modules/roleplay/api"
	"github.com/bwmarrin/discordgo"
)

func SwitchAPIHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var msg string
		if api.GetBAPI() == 1 {
			api.SetBAPI(0)
			msg = "Switched to OtakuGIFs API."
		} else {
			api.SetBAPI(1)
			msg = "Switched to Bastiwood API."
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: msg},
		})
	}
}
