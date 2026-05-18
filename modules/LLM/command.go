package LLM

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

func registerCommands() {
	cmd := &core.Command{
		Name:        "llm",
		Description: "Manage the LLM",
		AllowAdmin:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "load",
				Description: "Load the LLM model",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "unload",
				Description: "Unload the LLM model",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "prompt",
				Description: "Set the system prompt",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "The system prompt",
						Required:    true,
					},
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			options := i.ApplicationCommandData().Options
			if len(options) == 0 {
				return respond(s, i, "Missing subcommand.")
			}
			sub := options[0]
			switch sub.Name {
			case "load":
				if err := loadLLM(); err != nil {
					return respond(s, i, "Failed to load: "+err.Error())
				}
				return respond(s, i, "LLM loaded.")
			case "unload":
				if err := unloadLLM(); err != nil {
					return respond(s, i, "Failed to unload: "+err.Error())
				}
				return respond(s, i, "LLM unloaded.")
			case "prompt":
				setprompt(sub.Options[0].StringValue())
				return respond(s, i, "Prompt set.")
			}
			return respond(s, i, "Unknown subcommand.")
		},
	}
	core.Register(cmd)
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}
