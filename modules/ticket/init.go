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
		},
		Handler: ticketCommandHandler,
	}
	_ = core.Register(cmd)
}
