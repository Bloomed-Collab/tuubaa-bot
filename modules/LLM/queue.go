package LLM

import "github.com/bwmarrin/discordgo"

var queue = make(chan queueItem, 100)

type queueItem struct {
	s         *discordgo.Session
	channelID string
	message   string
}

// startWorker processes queued messages one at a time.
func startWorker() {
	go func() {
		for item := range queue {
			reply, err := getmessage(item.message)
			if err != nil {
				item.s.ChannelMessageSend(item.channelID, "Error: "+err.Error())
				continue
			}
			item.s.ChannelMessageSend(item.channelID, reply)
		}
	}()
}
