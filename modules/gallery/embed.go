package gallery

import (
	"fmt"
	"strings"
	"time"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

const accentColor = 0x9b59b6

func buildGalleryContainer(content string, imageURLs []string, timestamp time.Time, messageLink string) discordgo.MessageComponent {
	date := timestamp.UTC().Format("02 Jan 2006 · 15:04 UTC")

	container := v2.NewContainerBuilder().SetAccentColor(accentColor)

	if content != "" {
		container.AddComponent(v2.NewTextDisplayBuilder().SetContent(content).Build())
	}

	media := v2.NewMediaGalleryBuilder()
	for _, url := range imageURLs {
		media.AddImageURL(url)
	}
	container.AddComponent(media.Build())

	footer := fmt.Sprintf("-# %s　·　[↗ Original](%s)", date, messageLink)
	container.AddComponent(v2.NewTextDisplayBuilder().SetContent(footer).Build())

	return container.Build()
}

func updateGalleryStarCount(s *discordgo.Session, msg *discordgo.Message, post *postEntry) {
	starCount := 0
	for _, reaction := range msg.Reactions {
		if reaction.Emoji.Name == "⭐" {
			starCount = reaction.Count
			break
		}
	}

	var imageURLs []string
	for _, a := range msg.Attachments {
		if strings.HasPrefix(a.ContentType, "image/") || strings.HasPrefix(a.ContentType, "video/mp4") || isMediaURL(a.URL) {
			imageURLs = append(imageURLs, a.URL)
		}
	}

	messageLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", post.GuildID, post.ChannelID, post.MessageID)
	starDisplay := v2.NewTextDisplayBuilder().SetContent(fmt.Sprintf("⭐ %d", starCount)).Build()
	container := buildGalleryContainer(msg.Content, imageURLs, msg.Timestamp, messageLink)

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    post.ThreadID,
		ID:         post.PostID,
		Components: &[]discordgo.MessageComponent{starDisplay, container},
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	})
	if err != nil {
		logger.Warn("gallery: updateStarCount on post %s: %v", post.PostID, err)
	}
}
