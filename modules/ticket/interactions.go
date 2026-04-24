package ticket

import (
	"fmt"

	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func ticketInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	if i.Member == nil || i.GuildID == "" {
		return
	}

	switch i.MessageComponentData().CustomID {
	case "ticket:kunst":
		openTicket(s, i, "kunst")
	case "ticket:support":
		openTicket(s, i, "support")
	case "ticket:fanarts":
		openTicket(s, i, "fanarts")
	case "ticket:report":
		openTicket(s, i, "report")
	case "ticket:claim":
		handleClaim(s, i)
	case "ticket:close":
		handleClose(s, i)
	case "ticket:close_confirm":
		handleCloseConfirm(s, i)
	case "ticket:kunst:confirm":
		handleKunstConfirm(s, i)
	}
}

func hasTeamRole(guildID string, member *discordgo.Member) bool {
	if member.Permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}
	teamRoleID, _ := cfg.GetRole(guildID, "team_role")
	if teamRoleID == "" {
		return false
	}
	for _, r := range member.Roles {
		if r == teamRoleID {
			return true
		}
	}
	return false
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func deferEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})
}

func editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}) //nolint:errcheck
}

func resolveDisplayName(s *discordgo.Session, guildID, userID string) string {
	member, err := s.GuildMember(guildID, userID)
	if err != nil || member == nil {
		return userID
	}
	if member.Nick != "" {
		return member.Nick
	}
	if member.User != nil {
		if member.User.GlobalName != "" {
			return member.User.GlobalName
		}
		return member.User.Username
	}
	return userID
}

func openTicket(s *discordgo.Session, i *discordgo.InteractionCreate, kind string) {
	deferEphemeral(s, i)

	ticketChannelID, err := cfg.GetChannel(i.GuildID, "ticket")
	if err != nil || ticketChannelID == "" {
		editResponse(s, i, "Ticket-Kanal nicht konfiguriert. Bitte einen Admin fragen.")
		return
	}

	panelCh, err := s.Channel(ticketChannelID)
	if err != nil || panelCh.ParentID == "" {
		editResponse(s, i, "Ticket-Kategorie konnte nicht gefunden werden.")
		return
	}
	categoryID := panelCh.ParentID

	teamRoleID, _ := cfg.GetRole(i.GuildID, "team_role")
	if teamRoleID == "" {
		editResponse(s, i, "Team-Rolle nicht konfiguriert. Bitte einen Admin fragen (`/config setrole`).")
		return
	}

	k := ticketKinds[kind]
	displayName := resolveDisplayName(s, i.GuildID, i.Member.User.ID)
	channelName := fmt.Sprintf("𐔌₊˚꒰%s꒱﹕%s•˚₊⋅", k.icon, displayName)

	permOverwrites := []*discordgo.PermissionOverwrite{
		{
			ID:   i.GuildID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			ID:    i.Member.User.ID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionAttachFiles,
		},
		{
			ID:    teamRoleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionManageMessages,
		},
	}

	ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             categoryID,
		PermissionOverwrites: permOverwrites,
	})
	if err != nil {
		logger.Warn("ticket: create channel: %v", err)
		editResponse(s, i, "Fehler beim Erstellen des Ticket-Kanals.")
		return
	}

	infoComponents := buildInfoComponents(kind, false, teamRoleID)
	posted, err := s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Components: infoComponents,
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	})
	if err != nil {
		logger.Warn("ticket: send info card: %v", err)
	}

	msgID := ""
	if posted != nil {
		msgID = posted.ID
	}
	if err := saveTicket(ticketEntry{
		GuildID:   i.GuildID,
		ChannelID: ch.ID,
		MessageID: msgID,
		UserID:    i.Member.User.ID,
		Kind:      kind,
	}); err != nil {
		logger.Warn("ticket: saveTicket: %v", err)
	}

	editResponse(s, i, fmt.Sprintf("✅ Dein Ticket wurde erstellt: <#%s>", ch.ID))
}

func handleClaim(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasTeamRole(i.GuildID, i.Member) {
		respondEphemeral(s, i, "Bruh du kannst dein eigenes Ticket nicht claimen xD. Nur das Team kann das.")
		return
	}

	deferEphemeral(s, i)

	t, err := getTicket(i.GuildID, i.ChannelID)
	if err != nil || t == nil {
		editResponse(s, i, "Ticket nicht in der Datenbank gefunden.")
		return
	}
	if t.ClaimedBy != "" {
		editResponse(s, i, "Dieses Ticket wurde bereits geclaimed.")
		return
	}

	newComponents := buildInfoComponents(t.Kind, true, "")
	s.ChannelMessageEditComplex(&discordgo.MessageEdit{ //nolint:errcheck
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Components: &newComponents,
	})

	s.ChannelMessageSendComplex(i.ChannelID, buildClaimMessage(i.Member.User.ID)) //nolint:errcheck

	if err := claimTicket(i.GuildID, i.ChannelID, i.Member.User.ID); err != nil {
		logger.Warn("ticket: claimTicket: %v", err)
	}

	editResponse(s, i, "Erfolgreich geclaimed!")
}

func handleClose(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	if hasTeamRole(i.GuildID, i.Member) {
		deleteTicket(i.GuildID, i.ChannelID)
		s.ChannelDelete(i.ChannelID) //nolint:errcheck
		return
	}

	s.ChannelMessageSendComplex(i.ChannelID, buildCloseConfirmMessage(i.Member.User.ID)) //nolint:errcheck
}

func handleCloseConfirm(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasTeamRole(i.GuildID, i.Member) {
		respondEphemeral(s, i, "Nur das Team darf das Ticket schließen.")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	deleteTicket(i.GuildID, i.ChannelID)
	s.ChannelDelete(i.ChannelID) //nolint:errcheck
}

func handleKunstConfirm(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasTeamRole(i.GuildID, i.Member) {
		respondEphemeral(s, i, "Nur das Team kann die Künstlerrolle vergeben.")
		return
	}

	deferEphemeral(s, i)

	t, err := getTicket(i.GuildID, i.ChannelID)
	if err != nil || t == nil {
		editResponse(s, i, "Ticket nicht gefunden.")
		return
	}

	artistRoleID, _ := cfg.GetRole(i.GuildID, "artist_role")
	if artistRoleID == "" {
		editResponse(s, i, "Artist-Rolle nicht konfiguriert. Bitte `/config setrole` nutzen.")
		return
	}

	member, err := s.GuildMember(i.GuildID, t.UserID)
	if err != nil {
		editResponse(s, i, "Member nicht gefunden.")
		return
	}
	for _, r := range member.Roles {
		if r == artistRoleID {
			editResponse(s, i, fmt.Sprintf("<@%s> hat die Künstlerrolle bereits!", t.UserID))
			return
		}
	}

	if err := s.GuildMemberRoleAdd(i.GuildID, t.UserID, artistRoleID); err != nil {
		logger.Warn("ticket: add artist role: %v", err)
		editResponse(s, i, "Fehler beim Vergeben der Rolle.")
		return
	}

	s.ChannelMessageSendComplex(i.ChannelID, buildKunstConfirmMessage(t.UserID)) //nolint:errcheck
	editResponse(s, i, "Rolle erfolgreich vergeben!")
}
