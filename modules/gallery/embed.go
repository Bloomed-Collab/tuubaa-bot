package gallery

import (
	"fmt"
	"time"

	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
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
