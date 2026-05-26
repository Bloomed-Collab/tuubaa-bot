package embed

import (
	"fmt"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
)

const (
	ColorBan    = 0xe74c3c
	ColorUnban  = 0x2ecc71
	ColorMute   = 0xe67e22
	ColorUnmute = 0x2ecc71
	ColorKick   = 0xf1c40f
	ColorError  = 0xe74c3c
	ColorInfo   = 0x3498db
)

func Error(msg string) *discordgo.InteractionResponseData {
	c := v2.NewContainerBuilder().SetAccentColor(ColorError)
	c.AddComponent(v2.NewTextDisplayBuilder().SetContent("## Fehler").Build())
	c.AddComponent(v2.NewTextDisplayBuilder().SetContent(msg).Build())
	return &discordgo.InteractionResponseData{
		Components: []discordgo.MessageComponent{c.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2 | discordgo.MessageFlagsEphemeral,
	}
}

func modResponse(title, body, thumbnailURL string, color int) *discordgo.InteractionResponseData {
	c := v2.NewContainerBuilder().SetAccentColor(color)
	if thumbnailURL != "" {
		sec := v2.NewSectionBuilder().
			AddComponent(v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## %s", title)).Build()).
			AddComponent(v2.NewTextDisplayBuilder().SetContent(body).Build()).
			SetAccessory(v2.NewThumbnailBuilder().SetURL(thumbnailURL).Build())
		c.AddComponent(sec.Build())
	} else {
		c.AddComponent(v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## %s", title)).Build())
		c.AddComponent(v2.NewTextDisplayBuilder().SetContent(body).Build())
	}
	return &discordgo.InteractionResponseData{
		Components: []discordgo.MessageComponent{c.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func BanSuccess(user, moderator, reason, avatarURL string) *discordgo.InteractionResponseData {
	body := fmt.Sprintf("**Nutzer:** %s\n**Moderator:** %s\n**Grund:** %s", user, moderator, reason)
	return modResponse("Nutzer gebannt", body, avatarURL, ColorBan)
}

func UnbanSuccess(userID, moderator, reason string) *discordgo.InteractionResponseData {
	body := fmt.Sprintf("**Nutzer:** <@%s>\n**Moderator:** %s\n**Grund:** %s", userID, moderator, reason)
	return modResponse("Nutzer entbannt", body, "", ColorUnban)
}

func MuteSuccess(user, moderator, reason, duration, endsAt, avatarURL string) *discordgo.InteractionResponseData {
	body := fmt.Sprintf("**Nutzer:** %s\n**Moderator:** %s\n**Dauer:** %s\n**Endet:** %s\n**Grund:** %s", user, moderator, duration, endsAt, reason)
	return modResponse("Nutzer gemutet", body, avatarURL, ColorMute)
}

func UnmuteSuccess(user, moderator, reason, avatarURL string) *discordgo.InteractionResponseData {
	body := fmt.Sprintf("**Nutzer:** %s\n**Moderator:** %s\n**Grund:** %s", user, moderator, reason)
	return modResponse("Nutzer entmutet", body, avatarURL, ColorUnmute)
}

func KickSuccess(user, moderator, reason, avatarURL string) *discordgo.InteractionResponseData {
	body := fmt.Sprintf("**Nutzer:** %s\n**Moderator:** %s\n**Grund:** %s", user, moderator, reason)
	return modResponse("Nutzer gekickt", body, avatarURL, ColorKick)
}

func DmMessage(title, serverName, moderator, reason, thumbnailURL string, color int, extra string) *discordgo.MessageSend {
	c := v2.NewContainerBuilder().SetAccentColor(color)
	if thumbnailURL != "" {
		sec := v2.NewSectionBuilder().
			AddComponent(v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## %s", title)).Build()).
			SetAccessory(v2.NewThumbnailBuilder().SetURL(thumbnailURL).Build())
		c.AddComponent(sec.Build())
	} else {
		c.AddComponent(v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## %s", title)).Build())
	}
	body := fmt.Sprintf("**Server:** %s\n**Moderator:** %s\n**Grund:** %s", serverName, moderator, reason)
	if extra != "" {
		body += "\n" + extra
	}
	c.AddComponent(v2.NewTextDisplayBuilder().SetContent(body).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{c.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}

func LogMessage(title, user, moderator, reason, avatarURL string, color int, extra string) *discordgo.MessageSend {
	c := v2.NewContainerBuilder().SetAccentColor(color)
	if avatarURL != "" {
		sec := v2.NewSectionBuilder().
			AddComponent(v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## %s", title)).Build()).
			SetAccessory(v2.NewThumbnailBuilder().SetURL(avatarURL).Build())
		c.AddComponent(sec.Build())
	} else {
		c.AddComponent(v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("## %s", title)).Build())
	}
	body := fmt.Sprintf("**Nutzer:** %s\n**Moderator:** %s\n**Grund:** %s", user, moderator, reason)
	if extra != "" {
		body += "\n" + extra
	}
	c.AddComponent(v2.NewTextDisplayBuilder().SetContent(body).Build())
	return &discordgo.MessageSend{
		Components: []discordgo.MessageComponent{c.Build()},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	}
}
