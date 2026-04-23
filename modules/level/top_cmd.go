package level

import (
	"bytes"
	"fmt"
	"strings"

	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

const topPageSize = 10

func topCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ulog.Warn("top cmd: starting...")

	flags := discordgo.MessageFlags(0)
	if channelID, chErr := cfg.GetChannelCached(i.GuildID, "bot"); chErr == nil && channelID != "" && i.ChannelID != channelID {
		flags |= discordgo.MessageFlagsEphemeral
	}
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: flags},
	}); err != nil {
		return err
	}

	page := 1
	opts := i.ApplicationCommandData().Options
	if len(opts) > 0 {
		page = int(opts[0].IntValue())
		if page < 1 {
			page = 1
		}
	}

	comps, imgBuf, err := buildTopPage(s, i, page)
	if err != nil {
		ulog.Warn("top cmd: build error: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: ptrStr("Fehler beim Laden der Rangliste"),
		})
		return nil
	}

	params := &discordgo.WebhookParams{
		Components: comps,
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
	if flags&discordgo.MessageFlagsEphemeral != 0 {
		params.Flags |= discordgo.MessageFlagsEphemeral
	}
	if imgBuf != nil {
		params.Files = []*discordgo.File{
			{Name: "awesome.png", ContentType: "image/png", Reader: imgBuf},
		}
	}

	if _, followErr := s.FollowupMessageCreate(i.Interaction, true, params); followErr != nil {
		ulog.Warn("top cmd: followup error: %v", followErr)
		msg := "Fehler beim Senden der Rangliste"
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return followErr
	}
	return nil
}

func topButtonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	parts := strings.SplitN(data.CustomID, ":", 2)
	if len(parts) != 2 || parts[0] != "top" {
		return
	}

	page := 1
	fmt.Sscanf(parts[1], "%d", &page)
	if page < 1 {
		page = 1
	}

	comps, imgBuf, err := buildTopPage(s, i, page)
	if err != nil {
		ulog.Warn("top button: build error: %v", err)
		return
	}

	respData := &discordgo.InteractionResponseData{
		Components: comps,
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
	if imgBuf != nil {
		respData.Files = []*discordgo.File{
			{Name: "awesome.png", ContentType: "image/png", Reader: imgBuf},
		}
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: respData,
	})
}

func buildTopPage(s *discordgo.Session, i *discordgo.InteractionCreate, page int) ([]discordgo.MessageComponent, *bytes.Buffer, error) {
	allEntries, err := getAllXP()
	if err != nil {
		return nil, nil, err
	}

	guildID := i.GuildID
	var filtered []leaderboardEntry
	globalRank := 0
	for _, e := range allEntries {
		m, merr := s.State.Member(guildID, e.UserID)
		if merr != nil || m == nil {
			continue
		}
		globalRank++
		name := e.UserID
		avatar := ""
		if m.User != nil {
			name = m.User.DisplayName()
			avatar = m.User.AvatarURL("128")
		}
		filtered = append(filtered, leaderboardEntry{
			Rank:        globalRank,
			UserID:      e.UserID,
			XP:          e.XP,
			DisplayName: name,
			AvatarURL:   avatar,
		})
	}

	totalPages := (len(filtered) + topPageSize - 1) / topPageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	callerID := ""
	if i.Member != nil && i.Member.User != nil {
		callerID = i.Member.User.ID
	}

	callerRank := "##"
	for _, e := range filtered {
		if e.UserID == callerID {
			callerRank = fmt.Sprintf("#%d", e.Rank)
			break
		}
	}

	start := (page - 1) * topPageSize
	end := start + topPageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	pageEntries := filtered[start:end]

	comps := buildTopComponents(callerID, page, totalPages, callerRank)
	imgBuf, imgErr := buildLeaderboardImage(pageEntries)
	if imgErr != nil {
		ulog.Warn("top: image build error: %v", imgErr)
		return comps, nil, nil
	}
	return comps, imgBuf, nil
}
