package handler

import "github.com/bwmarrin/discordgo"

// AIQueue buffers incoming message requests so they are processed serially.
var AIQueue = make(chan QueueItem, 100)

// QueueItem holds everything needed to generate and send an AI reply.
type QueueItem struct {
	S         *discordgo.Session
	ChannelID string
	MessageID string
	Message   string
}

// StartWorker starts the background goroutine that drains AIQueue.
func StartWorker() {
	go func() {
		for item := range AIQueue {
			processItem(item)
		}
	}()
}

func processItem(item QueueItem) {
	var (
		reply string
		err   error
	)

	if ActiveBackend == BackendOpenAI {
		reply, err = GetAIResponseWithHistory(item.Message, item.ChannelID)
		if err == nil {
			AddMessageToCache(item.ChannelID, "assistant", reply)
		}
	} else {
		reply, err = GetBastiResponse(item.Message)
	}

	ref := &discordgo.MessageReference{
		MessageID: item.MessageID,
		ChannelID: item.ChannelID,
	}
	allowed := &discordgo.MessageAllowedMentions{RepliedUser: true}

	if err != nil {
		_, _ = item.S.ChannelMessageSendComplex(item.ChannelID, &discordgo.MessageSend{
			Content:         "❌ Fehler bei der Verarbeitung: " + err.Error(),
			Reference:       ref,
			AllowedMentions: allowed,
		})
		return
	}
	if len(reply) > 2000 {
		reply = reply[:1997] + "..."
	}
	_, _ = item.S.ChannelMessageSendComplex(item.ChannelID, &discordgo.MessageSend{
		Content:         reply,
		Reference:       ref,
		AllowedMentions: allowed,
	})
}
