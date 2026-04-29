package booster

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/booster/command"
	"github.com/S42yt/tuubaa-bot/modules/booster/event"
	"github.com/bwmarrin/discordgo"
)

func init() {
	options := []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "wahl",
			Description: "Wähle deine Rolle/Farbe",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Unschuldiges Kind", Value: "Unschuldiges Kind"},
				{Name: "Verdächtiges Kind", Value: "Verdächtiges Kind"},
				{Name: "Schuldiges Kind", Value: "Schuldiges Kind"},
				{Name: "Mit Entführer", Value: "Mit Entführer"},
				{Name: "Meisterentführer", Value: "Meisterentführer"},
				{Name: "Beifahrer", Value: "Beifahrer"},
				{Name: "Van Upgrader", Value: "Van Upgrader"},
				{Name: "Rainbow", Value: "Rainbow"},
			},
		},
	}

	farbenCmd := &core.Command{
		Name:          "farben",
		Description:   "Wähle eine Rolle/Farbe (nur mit Booster)",
		Options:       options,
		AllowEveryone: true,
		Handler:       command.FarbenHandler(),
	}

	_ = core.Register(farbenCmd)

	core.On(func(s *discordgo.Session, r *discordgo.Ready) {
		event.StartRainbowLoop(s)
	})
}
