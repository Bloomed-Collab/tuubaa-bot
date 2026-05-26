package roleplay

import (
	"math/rand"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/roleplay/commands"
	"github.com/bwmarrin/discordgo"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	subcommands := []struct {
		Name        string
		Description string
	}{
		{"cry", "Cry reaction"},
		{"pat", "Pat someone"},
		{"sad", "Sad reaction"},
		{"scared", "Scared reaction"},
		{"shy", "Shy reaction"},
		{"sleep", "Sleep reaction"},
		{"smug", "Smug reaction"},
		{"yay", "Yay reaction"},
		{"cuddle", "Cuddle reaction"},
		{"nervous", "Nervous reaction"},
		{"no", "No reaction"},
		{"cheers", "Cheers reaction"},
		{"blush", "Blush reaction"},
		{"slap", "Slap reaction"},
		{"cool", "Cool reaction"},
		{"hug", "Hug reaction"},
		{"facepalm", "Facepalm reaction"},
		{"happy", "Happy reaction"},
		{"laugh", "Laugh reaction"},
		{"mad", "Mad reaction"},
		{"love", "Love reaction"},
	}

	var options []*discordgo.ApplicationCommandOption
	var reactionChoices []*discordgo.ApplicationCommandOptionChoice
	for _, sc := range subcommands {
		options = append(options, &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        sc.Name,
			Description: sc.Description,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Optional target user",
					Required:    false,
				},
			},
		})
		reactionChoices = append(reactionChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  sc.Name,
			Value: sc.Name,
		})
	}

	rpCmd := &core.Command{
		Name:          "rp",
		Description:   "Roleplay reactions (subcommands)",
		Options:       options,
		AllowEveryone: true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 {
				return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: "Bitte Unterbefehl angeben."},
				})
			}
			sub := data.Options[0]
			return commands.RolePlayHandler(sub.Name)(s, i)
		},
	}

	_ = core.Register(rpCmd)

	cookieCmd := &core.Command{
		Name:        "cookie",
		Description: "Schenk jemanden einen Cookie",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Wem willst du einen Cookie geben?",
				Required:    true,
			},
		},
		AllowEveryone: true,
		Handler:       commands.CookieHandler(),
	}

	_ = core.Register(cookieCmd)

	switchAPICmd := &core.Command{
		Name:          "switchapi",
		Description:   "Switch the roleplay GIF API between OtakuGIFs and Bastiwood (Admin only)",
		AllowAdmin:    true,
		AllowEveryone: false,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "api",
				Description: "Which API to use",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Otaku", Value: "Otaku"},
					{Name: "Basti", Value: "Basti"},
					{Name: "Both", Value: "Both"},
				},
			},
		},
		Handler: commands.SwitchAPIHandler(),
	}

	_ = core.Register(switchAPICmd)

	setGifCmd := &core.Command{
		Name:        "setgif",
		Description: "Set GIF URL for a roleplay reaction in Basti API",
		AllowAdmin:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reaction",
				Description: "Pick one reaction",
				Required:    true,
				Choices:     reactionChoices,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "GIF URL to store",
				Required:    true,
			},
		},
		Handler: commands.SetGifHandler(),
	}

	_ = core.Register(setGifCmd)
}
