package gallery

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

func galleryCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(i.ApplicationCommandData().Options) == 0 {
		return respondEphemeral(s, i, "Unknown subcommand.")
	}

	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "rename":
		return handleRename(s, i, sub)
	}
	return respondEphemeral(s, i, "Unknown subcommand.")
}

func handleRename(s *discordgo.Session, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) error {
	newName := sub.Options[0].StringValue()
	if newName == "" {
		return respondEphemeral(s, i, "Name cannot be empty.")
	}

	threadID, err := getThread(i.GuildID, i.Member.User.ID)
	if err != nil {
		return respondEphemeral(s, i, "Could not look up your gallery.")
	}
	if threadID == "" {
		return respondEphemeral(s, i, "You don't have a gallery yet. Star one of your images in an art channel first.")
	}

	if _, err := s.ChannelEditComplex(threadID, &discordgo.ChannelEdit{Name: newName}); err != nil {
		return respondEphemeral(s, i, "Failed to rename your gallery: "+err.Error())
	}

	return respondEphemeral(s, i, "Deine gallery heißt jetzt **"+newName+"**.")
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func registerGalleryCommand() {
	cmd := &core.Command{
		Name:        "gallery",
		Description: "Manage your gallery",
		AllowEveryone: true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "rename",
				Description: "Rename your gallery thread",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "The new name for your gallery",
						Required:    true,
					},
				},
			},
		},
		Handler: galleryCommandHandler,
	}
	_ = core.Register(cmd)
}
