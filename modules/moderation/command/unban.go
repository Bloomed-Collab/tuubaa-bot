package command

import (
	modembed "github.com/S42yt/tuubaa-bot/modules/moderation/embed"
	"github.com/bwmarrin/discordgo"
)

func UnbanHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options[0].Options
		userOpt := getOption(opts, "user")
		if userOpt == nil {
			return respondError(s, i, "Du musst einen Nutzer angeben.")
		}
		target := userOpt.UserValue(s)
		reason := getReason(opts)
		mod := i.Member.User

		if err := s.GuildBanDelete(i.GuildID, target.ID); err != nil {
			return respondError(s, i, "Nutzer konnte nicht entbannt werden: "+err.Error())
		}

		sendLog(s, i.GuildID, modembed.LogMessage(
			"Mitglied entbannt",
			userTag(target), mod.Username, reason,
			avatarURL(target), modembed.ColorUnban, ""))

		return respond(s, i, modembed.UnbanSuccess(
			target.ID, mod.Username, reason))
	}
}
