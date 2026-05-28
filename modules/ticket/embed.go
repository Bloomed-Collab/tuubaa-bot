package ticket

import (
	"fmt"
	"time"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
)

var berlinLoc, _ = time.LoadLocation("Europe/Berlin")

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

	addUserBtn := discordgo.Button{
		Label:    "User hinzufügen",
		Style:    discordgo.SuccessButton,
		CustomID: "ticket:add_user",
		Emoji:    &discordgo.ComponentEmoji{Name: "➕"},
	}
	removeUserBtn := discordgo.Button{
		Label:    "User entfernen",
		Style:    discordgo.DangerButton,
		CustomID: "ticket:remove_user",
		Emoji:    &discordgo.ComponentEmoji{Name: "➖"},
	}

	var btns []discordgo.MessageComponent
	if kind == "kunst" {
		btns = []discordgo.MessageComponent{
			discordgo.Button{Label: "Accept", Style: discordgo.SuccessButton, CustomID: "ticket:kunst:confirm", Emoji: &discordgo.ComponentEmoji{Name: "✅"}},
			claimBtn,
			closeBtn,
		}
	} else if kind == "fanarts" {
		btns = []discordgo.MessageComponent{closeBtn}
	} else {
		btns = []discordgo.MessageComponent{claimBtn, closeBtn}
	}

	return []discordgo.MessageComponent{
		container.Build(),
		discordgo.ActionsRow{Components: btns},
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{addUserBtn, removeUserBtn}},
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

func buildUserAddedMessage(userID, addedByID string) *discordgo.MessageSend {
	container := v2.NewContainerBuilder().SetAccentColor(0x2ecc71)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("➕ <@%s> wurde von <@%s> zum Ticket hinzugefügt.", userID, addedByID),
	).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{container.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func buildUserRemovedMessage(userID, removedByID string) *discordgo.MessageSend {
	container := v2.NewContainerBuilder().SetAccentColor(0xe74c3c)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("➖ <@%s> wurde von <@%s> aus dem Ticket entfernt.", userID, removedByID),
	).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{container.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func buildUserSelectMessage(action string) *discordgo.InteractionResponse {
	var placeholder string
	if action == "add" {
		placeholder = "User auswählen zum Hinzufügen..."
	} else {
		placeholder = "User auswählen zum Entfernen..."
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "ticket:select_user:" + action,
							MenuType:    discordgo.UserSelectMenu,
							Placeholder: placeholder,
						},
					},
				},
			},
		},
	}
}

func buildTicketLogMessage(t *ticketEntry, openedByName, closedByName, closedByID string, closedAt time.Time) *discordgo.MessageSend {
	k := ticketKinds[t.Kind]
	if k.title == "" {
		k = ticketKinds["support"]
	}

	container := v2.NewContainerBuilder().SetAccentColor(k.color)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf("## 📋 Ticket Log — %s %s", k.icon, k.title),
	).Build())
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		fmt.Sprintf(
			"**Geöffnet von:** <@%s> (`%s`)\n"+
				"**Geöffnet am:** %s\n"+
				"**Geschlossen von:** <@%s> (`%s`)\n"+
				"**Geschlossen am:** %s",
			t.UserID, openedByName,
			t.OpenedAt.In(berlinLoc).Format("02.01.2006 um 15:04 Uhr"),
			closedByID, closedByName,
			closedAt.In(berlinLoc).Format("02.01.2006 um 15:04 Uhr"),
		),
	).Build())
	if t.ClaimedBy != "" {
		container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
			fmt.Sprintf("**Geclaimed von:** <@%s>", t.ClaimedBy),
		).Build())
	}
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(
		"-# Die Transcript-Dateien findest du unten angehängt (.txt & .html)",
	).Build())

	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{container.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}
