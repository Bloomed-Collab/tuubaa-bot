package command

import (
	modembed "github.com/S42yt/tuubaa-bot/modules/moderation/embed"
	"github.com/bwmarrin/discordgo"
)

func KickHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
			"Du wurdest gekickt",
			serverName, mod.Username, reason, iconURL,
			modembed.ColorKick, "Du kannst dem Server erneut beitreten."))

		if err := s.GuildMemberDeleteWithReason(i.GuildID, target.ID, reason); err != nil {
			return respondError(s, i, "Nutzer konnte nicht gekickt werden: "+err.Error())
		}

		sendLog(s, i.GuildID, modembed.LogMessage(
			"Mitglied gekickt",
			userTag(target), mod.Username, reason,
			avatarURL(target), modembed.ColorKick, ""))

		return respond(s, i, modembed.KickSuccess(
			userTag(target), mod.Username, reason, avatarURL(target)))
	}
}
