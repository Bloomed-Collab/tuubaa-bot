package level

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

var lvlMaxFloat = float64(lvlMax)

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

	setLevelCmd := &core.Command{
		Name:        "setlevel",
		Description: "Setzt das Level eines Users direkt",
		AllowAdmin:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Der User dessen Level gesetzt werden soll",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "level",
				Description: "Level (0–100)",
				Required:    true,
				MinValue:    floatPtr(0),
				MaxValue:    lvlMaxFloat,
			},
		},
		Handler: setLevelHandler,
	}

	syncCmd := &core.Command{
		Name:        "synclevels",
		Description: "Vergibt allen Mitgliedern ihre korrekte Level-Rolle basierend auf ihrem XP",
		AllowAdmin:  true,
		Handler:     syncLevelRolesHandler,
	}

	_ = core.Register(levelCmd)
	_ = core.Register(topCmd)
	_ = core.Register(setLevelCmd)
	_ = core.Register(syncCmd)

	core.On(topButtonHandler)
}

func floatPtr(f float64) *float64 { return &f }
