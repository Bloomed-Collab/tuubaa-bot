package embed

import (
	"fmt"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
)

func BuildRoleSetResponse(roleKey, roleID, executor string) *discordgo.InteractionResponseData {
	main := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## Configuration updated: %s", roleKey)).Build()
	footer := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("- Set to <@&%s> by %s", roleID, executor)).Build()

	comp := v2.NewContainerBuilder().SetAccentColor(0x99EE99).
		AddComponent(main).
		AddComponent(footer).
		Build()

	return &discordgo.InteractionResponseData{
		Components: []discordgo.MessageComponent{comp},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func BuildLevelRoleSetResponse(level, roleID, executor string) *discordgo.InteractionResponseData {
	main := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## Level Role gesetzt: Level %s", level)).Build()
	footer := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("- <@&%s> wird ab Level %s vergeben (gesetzt von %s)", roleID, level, executor)).Build()

	comp := v2.NewContainerBuilder().SetAccentColor(0x5865F2).
		AddComponent(main).
		AddComponent(footer).
		Build()

	return &discordgo.InteractionResponseData{
		Components: []discordgo.MessageComponent{comp},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func BuildChannelSetResponse(channelID, executor string, configChannel string) *discordgo.InteractionResponseData {
	main := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## Channel %s updated", configChannel)).Build()
	footer := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("- Set to <#%s> by %s", channelID, executor)).Build()

	comp := v2.NewContainerBuilder().SetAccentColor(0xFFD27A).
		AddComponent(main).
		AddComponent(footer).
		Build()

	return &discordgo.InteractionResponseData{
		Components: []discordgo.MessageComponent{comp},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}
