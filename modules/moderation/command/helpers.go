package command

import (
	"fmt"

	"github.com/S42yt/tuubaa-bot/modules/config"
	modembed "github.com/S42yt/tuubaa-bot/modules/moderation/embed"
	"github.com/bwmarrin/discordgo"
)

func resolveUser(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) (*discordgo.User, error) {
	for _, opt := range opts {
		if opt.Name == "user" {
			return opt.UserValue(s), nil
		}
	}
	for _, opt := range opts {
		if opt.Name == "username" {
			name := opt.StringValue()
			guild, err := s.Guild(i.GuildID)
			if err != nil {
				return nil, fmt.Errorf("server konnte nicht geladen werden")
			}
			members, err := s.GuildMembersSearch(guild.ID, name, 1)
			if err != nil || len(members) == 0 {
				return nil, fmt.Errorf("nutzer **%s** nicht gefunden", name)
			}
			return members[0].User, nil
		}
	}
	return nil, fmt.Errorf("kein Nutzer angegeben")
}

func getOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, o := range opts {
		if o.Name == name {
			return o
		}
	}
	return nil
}

func getReason(opts []*discordgo.ApplicationCommandInteractionDataOption) string {
	if o := getOption(opts, "reason"); o != nil {
		v := o.StringValue()
		if v != "" {
			return v
		}
	}
	return "Kein Grund angegeben"
}

func userTag(u *discordgo.User) string {
	return fmt.Sprintf("%s (<@%s>)", u.Username, u.ID)
}

func avatarURL(u *discordgo.User) string {
	return u.AvatarURL("256")
}

func guildIconURL(s *discordgo.Session, guildID string) string {
	guild, err := s.Guild(guildID)
	if err != nil || guild.Icon == "" {
		return ""
	}
	return discordgo.EndpointGuildIcon(guildID, guild.Icon)
}

func sendDM(s *discordgo.Session, userID string, msg *discordgo.MessageSend) {
	ch, err := s.UserChannelCreate(userID)
	if err != nil {
		return
	}
	s.ChannelMessageSendComplex(ch.ID, msg)
}

func sendLog(s *discordgo.Session, guildID string, msg *discordgo.MessageSend) {
	logCh, err := config.GetChannelCached(guildID, "mod_log")
	if err != nil || logCh == "" {
		logCh, err = config.GetChannelCached(guildID, "logs")
		if err != nil || logCh == "" {
			return
		}
	}
	s.ChannelMessageSendComplex(logCh, msg)
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return respond(s, i, modembed.Error(msg))
}
