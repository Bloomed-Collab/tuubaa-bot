package config

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/config/commands"
	"github.com/bwmarrin/discordgo"
)

func init() {
	setRole := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "setrole",
		Description: "Set a configured role for this guild",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "role",
				Description: "Which configurable role to set",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Unschuldiges Kind", Value: "ROLE_UNSCHULDIGES_KIND"},
					{Name: "Verdächtiges Kind", Value: "ROLE_VERDAECHTIGES_KIND"},
					{Name: "Schuldiges Kind", Value: "ROLE_SCHULDIGES_KIND"},
					{Name: "Mit Entführer", Value: "ROLE_MIT_ENTFUEHRER"},
					{Name: "Meisterentführer", Value: "ROLE_MEISTERENTFUEHRER"},
					{Name: "Beifahrer", Value: "ROLE_BEIFAHRER"},
					{Name: "Van Upgrader", Value: "ROLE_VAN_UPGRADER"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "target",
				Description: "The Discord role to assign for this key",
				Required:    true,
			},
		},
	}

	setChannel := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "setchannel",
		Description: "Set a configured channel for this guild",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "which",
				Description: "Which channel config to set",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Welcome Channel", Value: "welcome"},
					{Name: "Main Channel", Value: "main"},
					{Name: "Counter Channel", Value: "counterchannel"},
					{Name: "Logs Channel", Value: "logs"},
					{Name: "Bot Channel", Value: "bot"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to use for the selected config",
				Required:    true,
			},
		},
	}

	setLevelRole := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "setlevelrole",
		Description: "Rolle festlegen, die bei einem Level-Meilenstein vergeben wird",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "level",
				Description: "Level-Meilenstein",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Level 20", Value: "20"},
					{Name: "Level 40", Value: "40"},
					{Name: "Level 60", Value: "60"},
					{Name: "Level 80", Value: "80"},
					{Name: "Level 100", Value: "100"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Rolle, die ab diesem Level vergeben wird",
				Required:    true,
			},
		},
	}

	cfgCmd := &core.Command{
		Name:        "config",
		Description: "Guild-specific configuration",
		Options:     []*discordgo.ApplicationCommandOption{setRole, setChannel, setLevelRole},
		AllowAdmin:  true,
		Handler:     commands.ConfigRoleHandler(),
	}

	_ = core.Register(cfgCmd)
}
