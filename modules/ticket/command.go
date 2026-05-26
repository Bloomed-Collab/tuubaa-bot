package ticket

import (
	"fmt"
	"time"

	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func ticketCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(i.ApplicationCommandData().Options) == 0 {
		return nil
	}
	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "setup":
		return handleSetup(s, i)
	case "create":
		return handleCreate(s, i, sub)
	case "add_user":
		return handleAddUser(s, i, sub)
	case "remove_user":
		return handleRemoveUser(s, i, sub)
	}
	return nil
}

func handleSetup(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	deferEphemeral(s, i)

	ticketChannelID, err := cfg.GetChannel(i.GuildID, "ticket")
	if err != nil || ticketChannelID == "" {
		editResponse(s, i, "Ticket-Kanal nicht konfiguriert. Nutze `/config setchannel` → Ticket Channel.")
		return nil
	}

	if _, err := s.ChannelMessageSendComplex(ticketChannelID, buildPanelMessage()); err != nil {
		logger.Warn("ticket: send panel: %v", err)
		editResponse(s, i, fmt.Sprintf("Fehler beim Senden des Panels: %v", err))
		return nil
	}

	editResponse(s, i, fmt.Sprintf("✅ Ticket-Panel in <#%s> gesendet!", ticketChannelID))
	return nil
}

func handleCreate(s *discordgo.Session, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) error {
	var targetUser *discordgo.User
	for _, opt := range sub.Options {
		if opt.Name == "member" {
			targetUser = opt.UserValue(s)
		}
	}
	if targetUser == nil {
		respondEphemeral(s, i, "Member nicht gefunden.")
		return nil
	}

	deferEphemeral(s, i)

	ticketChannelID, err := cfg.GetChannel(i.GuildID, "ticket")
	if err != nil || ticketChannelID == "" {
		editResponse(s, i, "Ticket-Kanal nicht konfiguriert.")
		return nil
	}
	panelCh, err := s.Channel(ticketChannelID)
	if err != nil || panelCh.ParentID == "" {
		editResponse(s, i, "Ticket-Kategorie nicht gefunden.")
		return nil
	}

	teamRoleID, _ := cfg.GetRole(i.GuildID, "team_role")
	if teamRoleID == "" {
		editResponse(s, i, "Team-Rolle nicht konfiguriert. Nutze `/config setrole`.")
		return nil
	}

	k := ticketKinds["support"]
	displayName := resolveDisplayName(s, i.GuildID, targetUser.ID)
	channelName := fmt.Sprintf("𐔌₊˚꒰%s꒱﹕%s•˚₊⋅", k.icon, displayName)

	permOverwrites := []*discordgo.PermissionOverwrite{
		{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
		{ID: targetUser.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionAttachFiles},
		{ID: teamRoleID, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionManageMessages},
	}

	ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             panelCh.ParentID,
		PermissionOverwrites: permOverwrites,
	})
	if err != nil {
		logger.Warn("ticket: create channel (admin): %v", err)
		editResponse(s, i, "Fehler beim Erstellen des Ticket-Kanals.")
		return nil
	}

	posted, err := s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Components: buildInfoComponents("support", false, teamRoleID),
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	})

	msgID := ""
	if posted != nil {
		msgID = posted.ID
	}
	if err != nil {
		logger.Warn("ticket: send info card (admin): %v", err)
	}
	saveTicket(ticketEntry{ //nolint:errcheck
		GuildID:   i.GuildID,
		ChannelID: ch.ID,
		MessageID: msgID,
		UserID:    targetUser.ID,
		Kind:      "support",
		OpenedAt:  time.Now(),
	})

	editResponse(s, i, fmt.Sprintf("✅ Ticket erstellt: <#%s>", ch.ID))
	return nil
}

func handleAddUser(s *discordgo.Session, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) error {
	deferEphemeral(s, i)

	t, err := getTicket(i.GuildID, i.ChannelID)
	if err != nil || t == nil {
		editResponse(s, i, "Dieser Kanal ist kein Ticket.")
		return nil
	}

	if !hasTeamRole(i.GuildID, i.Member) {
		editResponse(s, i, "Nur das Team kann User zu Tickets hinzufügen.")
		return nil
	}

	var targetUser *discordgo.User
	for _, opt := range sub.Options {
		if opt.Name == "member" {
			targetUser = opt.UserValue(s)
		}
	}
	if targetUser == nil {
		editResponse(s, i, "Member nicht gefunden.")
		return nil
	}

	err = s.ChannelPermissionSet(i.ChannelID, targetUser.ID, discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionAttachFiles, 0)
	if err != nil {
		logger.Warn("ticket: add user perm: %v", err)
		editResponse(s, i, "Fehler beim Hinzufügen des Users.")
		return nil
	}

	s.ChannelMessageSendComplex(i.ChannelID, buildUserAddedMessage(targetUser.ID, i.Member.User.ID)) //nolint:errcheck
	editResponse(s, i, fmt.Sprintf("✅ <@%s> wurde zum Ticket hinzugefügt.", targetUser.ID))
	return nil
}

func handleRemoveUser(s *discordgo.Session, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) error {
	deferEphemeral(s, i)

	t, err := getTicket(i.GuildID, i.ChannelID)
	if err != nil || t == nil {
		editResponse(s, i, "Dieser Kanal ist kein Ticket.")
		return nil
	}

	if !hasTeamRole(i.GuildID, i.Member) {
		editResponse(s, i, "Nur das Team kann User aus Tickets entfernen.")
		return nil
	}

	var targetUser *discordgo.User
	for _, opt := range sub.Options {
		if opt.Name == "member" {
			targetUser = opt.UserValue(s)
		}
	}
	if targetUser == nil {
		editResponse(s, i, "Member nicht gefunden.")
		return nil
	}

	if targetUser.ID == t.UserID {
		editResponse(s, i, "Du kannst den Ticket-Ersteller nicht entfernen.")
		return nil
	}

	err = s.ChannelPermissionDelete(i.ChannelID, targetUser.ID)
	if err != nil {
		logger.Warn("ticket: remove user perm: %v", err)
		editResponse(s, i, "Fehler beim Entfernen des Users.")
		return nil
	}

	s.ChannelMessageSendComplex(i.ChannelID, buildUserRemovedMessage(targetUser.ID, i.Member.User.ID)) //nolint:errcheck
	editResponse(s, i, fmt.Sprintf("✅ <@%s> wurde aus dem Ticket entfernt.", targetUser.ID))
	return nil
}
