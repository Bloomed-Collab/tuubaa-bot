package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	modembed "github.com/S42yt/tuubaa-bot/modules/moderation/embed"
	"github.com/bwmarrin/discordgo"
)

func parseDuration(input string) (time.Duration, string, error) {
	parts := strings.Split(input, ".")
	if len(parts) != 3 {
		return 0, "", fmt.Errorf("ungültiges Format, nutze `d.h.m` (z.B. `0.1.30`)")
	}
	days, err := strconv.Atoi(parts[0])
	if err != nil || days < 0 {
		return 0, "", fmt.Errorf("ungültiger Tage-Wert")
	}
	hours, err := strconv.Atoi(parts[1])
	if err != nil || hours < 0 {
		return 0, "", fmt.Errorf("ungültiger Stunden-Wert")
	}
	minutes, err := strconv.Atoi(parts[2])
	if err != nil || minutes < 0 {
		return 0, "", fmt.Errorf("ungültiger Minuten-Wert")
	}
	d := time.Duration(days)*24*time.Hour + time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	if d <= 0 {
		return 0, "", fmt.Errorf("dauer muss größer als 0 sein")
	}
	if d > 28*24*time.Hour {
		return 0, "", fmt.Errorf("maximale Mute-Dauer ist 28 Tage")
	}
	var label []string
	if days > 0 {
		label = append(label, fmt.Sprintf("%d Tag(e)", days))
	}
	if hours > 0 {
		label = append(label, fmt.Sprintf("%d Stunde(n)", hours))
	}
	if minutes > 0 {
		label = append(label, fmt.Sprintf("%d Minute(n)", minutes))
	}
	return d, strings.Join(label, " "), nil
}

func MuteHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		opts := i.ApplicationCommandData().Options[0].Options
		target, err := resolveUser(s, i, opts)
		if err != nil {
			return respondError(s, i, err.Error())
		}

		lengthOpt := getOption(opts, "length")
		if lengthOpt == nil {
			return respondError(s, i, "Du musst eine Dauer angeben (Format: `d.h.m`).")
		}

		dur, label, err := parseDuration(lengthOpt.StringValue())
		if err != nil {
			return respondError(s, i, err.Error())
		}

		reason := getReason(opts)
		mod := i.Member.User
		until := time.Now().Add(dur)
		untilStr := fmt.Sprintf("<t:%d:F>", until.Unix())

		timeoutUntil := until
		if err := s.GuildMemberTimeout(i.GuildID, target.ID, &timeoutUntil); err != nil {
			return respondError(s, i, "Nutzer konnte nicht gemutet werden: "+err.Error())
		}

		iconURL := guildIconURL(s, i.GuildID)
		guild, _ := s.Guild(i.GuildID)
		serverName := i.GuildID
		if guild != nil {
			serverName = guild.Name
		}

		sendDM(s, target.ID, modembed.DmMessage(
			"Du wurdest gemutet",
			serverName, mod.Username, reason, iconURL,
			modembed.ColorMute,
			fmt.Sprintf("**Dauer:** %s\n**Endet:** %s", label, untilStr)))

		sendLog(s, i.GuildID, modembed.LogMessage(
			"Mitglied gemutet",
			userTag(target), mod.Username, reason,
			avatarURL(target), modembed.ColorMute,
			fmt.Sprintf("**Dauer:** %s\n**Endet:** %s", label, untilStr)))

		return respond(s, i, modembed.MuteSuccess(
			userTag(target), mod.Username, reason, label, untilStr, avatarURL(target)))
	}
}
