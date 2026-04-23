package level

import (
	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func levelCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ulog.Warn("level cmd: starting...")

	// Respond ASAP to avoid "Unknown interaction" if DB is slow.
	flags := discordgo.MessageFlags(0)
	if channelID, chErr := cfg.GetChannelCached(i.GuildID, "bot"); chErr == nil && channelID != "" && i.ChannelID != channelID {
		flags |= discordgo.MessageFlagsEphemeral
	}
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: flags},
	}); err != nil {
		ulog.Warn("level cmd: defer respond error: %v", err)
		return err
	}

	var targetUser *discordgo.User
	opts := i.ApplicationCommandData().Options
	if len(opts) > 0 && opts[0].Type == discordgo.ApplicationCommandOptionUser {
		targetUser = opts[0].UserValue(s)
	} else {
		targetUser = i.Member.User
		if targetUser == nil {
			targetUser = i.User
		}
	}

	if targetUser == nil {
		ulog.Warn("level cmd: could not determine target user")
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: ptrStr("Fehler: Benutzer konnte nicht ermittelt werden"),
		})
		return nil
	}

	xp, _ := getXP(targetUser.ID)
	all, _ := getAllXP()

	rank := len(all) + 1
	for idx, e := range all {
		if e.UserID == targetUser.ID {
			rank = idx + 1
			break
		}
	}

	level := calcLevel(xp)
	fromLevel := xpFromThisLevel(xp)
	toNext := xpToNextLevel(xp)
	requiredForLevel := fromLevel + toNext

	displayName := targetUser.DisplayName()
	avatarURL := targetUser.AvatarURL("256")

	imgBuf, imgErr := buildRankCard(displayName, avatarURL, level, rank, fromLevel, requiredForLevel)
	if imgErr != nil {
		ulog.Warn("level cmd: rank card error: %v", imgErr)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: ptrStr("Fehler beim Erstellen der Levelkarte"),
		})
		return imgErr
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files: []*discordgo.File{
			{Name: "level.png", ContentType: "image/png", Reader: imgBuf},
		},
	})
	if err != nil {
		ulog.Warn("level cmd: response edit error: %v", err)
		fallback := "Fehler beim Senden der Levelkarte"
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &fallback,
		})
		return err
	}
	return nil
}

func ptrStr(s string) *string {
	return &s
}
