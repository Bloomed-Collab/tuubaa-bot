package chatgpt

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/chatgpt/commands"
	_ "github.com/S42yt/tuubaa-bot/modules/chatgpt/events"
	"github.com/S42yt/tuubaa-bot/modules/chatgpt/handler"
	"github.com/bwmarrin/discordgo"
)

func init() {
	handler.StartWorker()

	_ = core.Register(&core.Command{
		Name:        "changeai",
		Description: "Switch the active AI backend",
		AllowAdmin:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "backend",
				Description: "The AI backend to activate",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "ChatGPT (OpenAI)", Value: "openai"},
					{Name: "Basti API", Value: "basti"},
					{Name: "Disabled", Value: "disabled"},
				},
			},
		},
		Handler: commands.ChangeAIHandler(),
	})

	_ = core.Register(&core.Command{
		Name:        "loadai",
		Description: "Load the Basti AI model",
		AllowAdmin:  true,
		Handler:     commands.LoadAIHandler(),
	})

	_ = core.Register(&core.Command{
		Name:        "unloadai",
		Description: "Unload the Basti AI model",
		AllowAdmin:  true,
		Handler:     commands.UnloadAIHandler(),
	})

	_ = core.Register(&core.Command{
		Name:        "setprompt",
		Description: "Set the system prompt for the Basti AI model",
		AllowAdmin:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "prompt",
				Description: "The system prompt text",
				Required:    true,
			},
		},
		Handler: commands.SetPromptHandler(),
	})
}
