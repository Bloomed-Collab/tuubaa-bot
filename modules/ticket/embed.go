package ticket

import (
	"fmt"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
)

type kindInfo struct {
	icon        string
	title       string
	description string
	color       int
}

var ticketKinds = map[string]kindInfo{
	"kunst": {
		icon:        "🎨",
		title:       "Antrag auf die Künstlerrolle",
		description: "Sobald du Level 10 erreicht hast, kannst du dich hier für die Künstlerrolle bewerben!\nMit `/level` kannst du dein aktuelles Level einsehen.",
		color:       0xe67e22,
	},
	"support": {
		icon:        "🔧",
		title:       "Support",
		description: "Das Team wird in kürze da sein.",
		color:       0x3498db,
	},
	"fanarts": {
		icon:        "🖼️",
		title:       "Fanart an tuubaa",
		description: "Sende hier deine Fanart an tuubaa.",
		color:       0xf1c40f,
	},
	"report": {
		icon:        "📢",
		title:       "Reporte eine Person",
		description: "Das Team wird in kürze da sein.",
		color:       0xe74c3c,
	},
}

func buildPanelMessage() *discordgo.MessageSend {
	container := v2.NewContainerBuilder().SetAccentColor(0x9b59b6)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		"## 🎫 Ticket System\n" +
			"Wähle eine Kategorie um ein Ticket zu öffnen.\n\n" +
			"🎨 **Kunst** -Beantrage die Künstlerrolle (ab Level 10)\n" +
			"🔧 **Support** -Brauchst du Hilfe? Erstelle hier ein Support Ticket!\n" +
			"🖼️ **Fanarts** -Schicke Fanarts an tuubaa\n" +
			"📢 **Report** -Melde eine Person",
	).Build())

	buttons := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{Label: "Kunst", Style: discordgo.SuccessButton, CustomID: "ticket:kunst", Emoji: &discordgo.ComponentEmoji{Name: "🎨"}},
			discordgo.Button{Label: "Support", Style: discordgo.SecondaryButton, CustomID: "ticket:support", Emoji: &discordgo.ComponentEmoji{Name: "🔧"}},
			discordgo.Button{Label: "Fanarts", Style: discordgo.PrimaryButton, CustomID: "ticket:fanarts", Emoji: &discordgo.ComponentEmoji{Name: "🖼️"}},
			discordgo.Button{Label: "Report", Style: discordgo.DangerButton, CustomID: "ticket:report", Emoji: &discordgo.ComponentEmoji{Name: "📢"}},
		},
	}

	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{container.Build(), buttons},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func buildInfoComponents(kind string, claimDisabled bool, teamRoleID string) []discordgo.MessageComponent {
	k, ok := ticketKinds[kind]
	if !ok {
		k = ticketKinds["support"]
	}

	container := v2.NewContainerBuilder().SetAccentColor(k.color)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("## %s %s\n%s", k.icon, k.title, k.description),
	).Build())
	if teamRoleID != "" {
		container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
			fmt.Sprintf("-# <@&%s>", teamRoleID),
		).Build())
	}

	claimBtn := discordgo.Button{
		Label:    "Claim",
		Style:    discordgo.PrimaryButton,
		CustomID: "ticket:claim",
		Disabled: claimDisabled,
		Emoji:    &discordgo.ComponentEmoji{Name: "🙋"},
	}
	closeBtn := discordgo.Button{
		Label:    "Schließen",
		Style:    discordgo.DangerButton,
		CustomID: "ticket:close",
		Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
	}

	var btns []discordgo.MessageComponent
	if kind == "kunst" {
		btns = []discordgo.MessageComponent{
			discordgo.Button{Label: "Accept", Style: discordgo.SuccessButton, CustomID: "ticket:kunst:confirm", Emoji: &discordgo.ComponentEmoji{Name: "✅"}},
			claimBtn,
			closeBtn,
		}
	} else {
		btns = []discordgo.MessageComponent{claimBtn, closeBtn}
	}

	return []discordgo.MessageComponent{
		container.Build(),
		discordgo.ActionsRow{Components: btns},
	}
}

func buildClaimMessage(userID string) *discordgo.MessageSend {
	container := v2.NewContainerBuilder().SetAccentColor(0x3498db)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("<@%s> wird sich nun um das Ticket kümmern ^^", userID),
	).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{container.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func buildCloseConfirmMessage(userID string) *discordgo.MessageSend {
	container := v2.NewContainerBuilder().SetAccentColor(0xe74c3c)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("<@%s> möchte das Ticket schließen", userID),
	).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{
			container.Build(),
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Schließen bestätigen",
						Style:    discordgo.DangerButton,
						CustomID: "ticket:close_confirm",
						Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
					},
				},
			},
		},
		Flags: discordgo.MessageFlagsIsComponentsV2,
	}
}

func buildKunstConfirmMessage(userID string) *discordgo.MessageSend {
	container := v2.NewContainerBuilder().SetAccentColor(0x2ecc71)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("<@%s>! Du hast erfolgreich die Künstler Rolle erhalten :D", userID),
	).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{container.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}
