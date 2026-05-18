package LLM

import "github.com/bwmarrin/discordgo"

var queue = make(chan queueItem, 100)

type queueItem struct {
	s         *discordgo.Session
	channelID string
	messageID string
	message   string
}

// startWorker processes queued messages one at a time.
func startWorker() {
	go func() {
		for item := range queue {
			reply, err := getmessage(item.message)
			if err != nil {
				_, _ = item.s.ChannelMessageSendComplex(item.channelID, &discordgo.MessageSend{
					Content: "Error: " + err.Error(),
					Reference: &discordgo.MessageReference{
						MessageID: item.messageID,
						ChannelID: item.channelID,
					},
					AllowedMentions: &discordgo.MessageAllowedMentions{RepliedUser: true},
				})
				continue
			}
			_, _ = item.s.ChannelMessageSendComplex(item.channelID, &discordgo.MessageSend{
				Content: reply,
				Reference: &discordgo.MessageReference{
					MessageID: item.messageID,
					ChannelID: item.channelID,
				},
				AllowedMentions: &discordgo.MessageAllowedMentions{RepliedUser: true},
			})
		}
	}()
}
