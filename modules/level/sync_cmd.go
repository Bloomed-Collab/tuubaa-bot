package level

import (
	"fmt"

	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func syncLevelRolesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	// Load all configured level roles up front
	roleMap := map[int]string{} // threshold → roleID
	for _, t := range levelRoleThresholds {
		roleID, err := cfg.GetLevelRole(i.GuildID, t)
		if err == nil && roleID != "" {
			roleMap[t] = roleID
		}
	}
	if len(roleMap) == 0 {
		content := "Keine Level-Rollen konfiguriert. Nutze `/config setlevelrole`."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}) //nolint:errcheck
		return nil
	}

	// All role IDs as a flat list for easy removal
	allRoleIDs := make([]string, 0, len(roleMap))
	for _, id := range roleMap {
		allRoleIDs = append(allRoleIDs, id)
	}

	entries, err := getAllXP()
	if err != nil {
		content := fmt.Sprintf("Fehler beim Laden der XP-Daten: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}) //nolint:errcheck
		return nil
	}

	var updated, noRole, notInGuild int

	for _, e := range entries {
		level := calcLevel(e.XP)
		targetThreshold := highestReachedThreshold(level)

		// Remove all level roles first
		for _, roleID := range allRoleIDs {
			s.GuildMemberRoleRemove(i.GuildID, e.UserID, roleID) //nolint:errcheck
		}

		if targetThreshold == 0 {
			noRole++
			continue
		}

		targetRoleID, ok := roleMap[targetThreshold]
		if !ok {
			noRole++
			continue
		}

		if err := s.GuildMemberRoleAdd(i.GuildID, e.UserID, targetRoleID); err != nil {
			ulog.Warn("synclevels: add role to %s: %v", e.UserID, err)
			notInGuild++
			continue
		}
		updated++
	}

	content := fmt.Sprintf(
		"✅ Sync abgeschlossen!\n\n**%d** Rollen vergeben\n**%d** User nicht im Server\n**%d** User unter Level 20",
		updated, notInGuild, noRole,
	)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}) //nolint:errcheck
	return nil
}

// highestReachedThreshold returns the highest configured threshold the given level has crossed.
// Returns 0 if below all thresholds.
func highestReachedThreshold(level int) int {
	highest := 0
	for _, t := range levelRoleThresholds {
		if level >= t && t > highest {
			highest = t
		}
	}
	return highest
}
