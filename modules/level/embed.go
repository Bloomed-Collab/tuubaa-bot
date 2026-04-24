package level

import (
	"fmt"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
)

func buildLevelUpComponents(displayName string, level int, roleName string) []discordgo.MessageComponent {
	var content string
	if roleName != "" {
		content = fmt.Sprintf("### %s hat Level %d erreicht und die Rolle **%s** erhalten! Glückwunsch!", displayName, level, roleName)
	} else {
		content = fmt.Sprintf("### %s hat Level %d erreicht!", displayName, level)
	}
	text := v2.NewTextDisplayBuilder().SetContent(content).Build()
	return []discordgo.MessageComponent{text}
}

func buildTopComponents(callerID string, page, totalPages int, callerRank string) []discordgo.MessageComponent {
	text := v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("# 🏆 Topliste\n**Seite %d / %d**\n<@%s> ist auf Platz %s", page, totalPages, callerID, callerRank),
	).Build()

	mg := v2.NewMediaGalleryBuilder().AddImageURL("attachment://awesome.png").Build()

	buttons := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
				CustomID: fmt.Sprintf("top:%d", page-1),
				Style:    discordgo.SecondaryButton,
				Disabled: page <= 1,
			},
			discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				CustomID: fmt.Sprintf("top:%d", page+1),
				Style:    discordgo.SecondaryButton,
				Disabled: page >= totalPages,
			},
		},
	}

	container := v2.NewContainerBuilder().
		SetAccentColor(0x5865F2).
		AddComponent(text).
		AddComponent(mg).
		AddComponent(buttons).
		Build()

	return []discordgo.MessageComponent{container}
}
