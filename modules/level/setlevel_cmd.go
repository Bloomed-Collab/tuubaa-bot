package level

import (
	"fmt"

	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func setLevelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	opts := i.ApplicationCommandData().Options
	var targetUser *discordgo.User
	var targetLevel int64

	for _, opt := range opts {
		switch opt.Name {
		case "user":
			targetUser = opt.UserValue(s)
		case "level":
			targetLevel = opt.IntValue()
		}
	}

	if targetUser == nil {
		return respond(s, i, "User nicht gefunden.")
	}
	if targetLevel < 0 || targetLevel > int64(lvlMax) {
		return respond(s, i, fmt.Sprintf("Level muss zwischen 0 und %d liegen.", lvlMax))
	}

	xp := totalXPForLevel(int(targetLevel))

	if err := upsertXP(targetUser.ID, xp); err != nil {
		ulog.Warn("setlevel: upsertXP(%s): %v", targetUser.ID, err)
		return respond(s, i, fmt.Sprintf("Fehler beim Setzen des Levels: %v", err))
	}

	return respond(s, i, fmt.Sprintf("<@%s> ist jetzt **Level %d** (%d XP).", targetUser.ID, targetLevel, xp))
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
