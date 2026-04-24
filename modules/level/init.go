package level

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

func init() {
	levelCmd := &core.Command{
		Name:          "level",
		Description:   "Zeigt dein Level oder das eines anderen Benutzers an",
		AllowEveryone: true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Benutzer (optional)",
				Required:    false,
			},
		},
		Handler: levelCommandHandler,
	}

	topCmd := &core.Command{
		Name:          "top",
		Description:   "Zeigt die Rangliste der aktivsten Mitglieder",
		AllowEveryone: true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "seite",
				Description: "Seite der Rangliste (Standard: 1)",
				Required:    false,
				MinValue:    floatPtr(1),
			},
		},
		Handler: topCommandHandler,
	}

	_ = core.Register(levelCmd)
	_ = core.Register(topCmd)

	core.On(topButtonHandler)
}

func floatPtr(f float64) *float64 { return &f }
