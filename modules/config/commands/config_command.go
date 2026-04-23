package commands

import (
	"github.com/bwmarrin/discordgo"
)

func ConfigRoleHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data := i.ApplicationCommandData()
		if len(data.Options) == 0 {
			return respond(s, i, "No subcommand provided")
		}

		sub := data.Options[0]
		switch sub.Name {
		case "setrole":
			return handleSetRole(s, i)
		case "setchannel":
			return handleSetChannel(s, i)
		case "setlevelrole":
			return handleSetLevelRole(s, i)
		default:
			return respond(s, i, "Unknown subcommand")
		}
	}
}
