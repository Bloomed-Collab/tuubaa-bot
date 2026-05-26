package command

import (
	modembed "github.com/S42yt/tuubaa-bot/modules/moderation/embed"
	"github.com/bwmarrin/discordgo"
)

func BanHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options[0].Options
		target, err := resolveUser(s, i, opts)
		if err != nil {
			return respondError(s, i, err.Error())
		}
		reason := getReason(opts)
		mod := i.Member.User

		iconURL := guildIconURL(s, i.GuildID)
		guild, _ := s.Guild(i.GuildID)
		serverName := i.GuildID
		if guild != nil {
			serverName = guild.Name
		}

		sendDM(s, target.ID, modembed.DmMessage(
			"Du wurdest gebannt",
			serverName, mod.Username, reason, iconURL,
			modembed.ColorBan, "Dieser Bann ist **permanent**."))

		if err := s.GuildBanCreateWithReason(i.GuildID, target.ID, reason, 1); err != nil {
			return respondError(s, i, "Nutzer konnte nicht gebannt werden: "+err.Error())
		}

		sendLog(s, i.GuildID, modembed.LogMessage(
			"Mitglied gebannt",
			userTag(target), mod.Username, reason,
			avatarURL(target), modembed.ColorBan, ""))

		return respond(s, i, modembed.BanSuccess(
			userTag(target), mod.Username, reason, avatarURL(target)))
	}
}
