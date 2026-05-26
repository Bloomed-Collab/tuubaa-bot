package ticket

import (
	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

func init() {
	core.On(ticketInteractionHandler)

	cmd := &core.Command{
		Name:        "ticket",
		Description: "Ticket-System verwalten",
		AllowAdmin:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "setup",
				Description: "Sendet das Ticket-Panel in den konfigurierten Ticket-Kanal",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "create",
				Description: "Erstelle manuell ein Ticket für ein Mitglied",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "member",
						Description: "Das Mitglied für das das Ticket erstellt wird",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add_user",
				Description: "Füge einen User zum aktuellen Ticket hinzu",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "member",
						Description: "Der User der hinzugefügt werden soll",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove_user",
				Description: "Entferne einen User aus dem aktuellen Ticket",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "member",
						Description: "Der User der entfernt werden soll",
						Required:    true,
					},
				},
			},
		},
		Handler: ticketCommandHandler,
	}
	_ = core.Register(cmd)
}
