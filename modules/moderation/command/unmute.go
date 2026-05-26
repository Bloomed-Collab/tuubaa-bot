package command

import (
	modembed "github.com/S42yt/tuubaa-bot/modules/moderation/embed"
	"github.com/bwmarrin/discordgo"
)

func UnmuteHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options[0].Options
		target, err := resolveUser(s, i, opts)
		if err != nil {
			return respondError(s, i, err.Error())
		}
		reason := getReason(opts)
		mod := i.Member.User

		if err := s.GuildMemberTimeout(i.GuildID, target.ID, nil); err != nil {
			return respondError(s, i, "Nutzer konnte nicht entmutet werden: "+err.Error())
		}

		iconURL := guildIconURL(s, i.GuildID)
		guild, _ := s.Guild(i.GuildID)
		serverName := i.GuildID
		if guild != nil {
			serverName = guild.Name
		}

		sendDM(s, target.ID, modembed.DmMessage(
			"Du wurdest entmutet",
			serverName, mod.Username, reason, iconURL,
			modembed.ColorUnmute, ""))

		sendLog(s, i.GuildID, modembed.LogMessage(
			"Mitglied entmutet",
			userTag(target), mod.Username, reason,
			avatarURL(target), modembed.ColorUnmute, ""))

		return respond(s, i, modembed.UnmuteSuccess(
			userTag(target), mod.Username, reason, avatarURL(target)))
	}
}
