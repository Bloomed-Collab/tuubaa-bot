package moderation

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/moderation/command"
	"github.com/bwmarrin/discordgo"
)

func init() {
	userOpt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        "user",
		Description: "Der User (Erwähnung oder ID)",
		Required:    false,
	}

	usernameOpt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "username",
		Description: "Nach Username suchen",
		Required:    false,
	}

	reasonOpt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "reason",
		Description: "Grund für die Aktion",
		Required:    false,
	}

	ban := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "ban",
		Description: "Einen User permanent vom Server bannen",
		Options: []*discordgo.ApplicationCommandOption{
			userOpt, usernameOpt, reasonOpt,
		},
	}

	unban := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "unban",
		Description: "Einen User entbannen",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Die User-ID zum Entbannen",
				Required:    true,
			},
			reasonOpt,
		},
	}

	mute := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "mute",
		Description: "Einen User für eine bestimmte Dauer muten",
		Options: []*discordgo.ApplicationCommandOption{
			userOpt, usernameOpt,
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "length",
				Description: "Dauer im Format d.h.m (z.B. 1.2.30 = 1 Tag 2 Stunden 30 Minuten)",
				Required:    true,
			},
			reasonOpt,
		},
	}

	unmute := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "unmute",
		Description: "Einen User entmuten",
		Options: []*discordgo.ApplicationCommandOption{
			userOpt, usernameOpt, reasonOpt,
		},
	}

	kick := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "kick",
		Description: "Einen User vom Server kicken",
		Options: []*discordgo.ApplicationCommandOption{
			userOpt, usernameOpt, reasonOpt,
		},
	}

	cmd := &core.Command{
		Name:        "mod",
		Description: "Moderations-Befehle",
		Options:     []*discordgo.ApplicationCommandOption{ban, unban, mute, unmute, kick},
		AllowAdmin:  true,
		Handler:     routeHandler(),
	}

	_ = core.Register(cmd)
}

func routeHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"ban":    command.BanHandler(),
		"unban":  command.UnbanHandler(),
		"mute":   command.MuteHandler(),
		"unmute": command.UnmuteHandler(),
		"kick":   command.KickHandler(),
	}

	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		sub := i.ApplicationCommandData().Options[0].Name
		if h, ok := handlers[sub]; ok {
			return h(s, i)
		}
		return nil
	}
}
