package handler

import (
	"sync"
	"time"

	"github.com/openai/openai-go"
)

type ConversationMessage struct {
	Role      string 
	Content   string
	Timestamp time.Time
}

type ChannelCache struct {
	Messages []*ConversationMessage
	Mu       sync.RWMutex
}

var (
	conversationCache = map[string]*ChannelCache{}
	cacheMu           sync.RWMutex
	cacheExpiry       = 5 * time.Minute
)

func AddMessageToCache(channelID string, role string, content string) {
	cacheMu.Lock()
	if _, exists := conversationCache[channelID]; !exists {
		conversationCache[channelID] = &ChannelCache{
			Messages: make([]*ConversationMessage, 0),
		}
	}
	cache := conversationCache[channelID]
	cacheMu.Unlock()

	cache.Mu.Lock()
	defer cache.Mu.Unlock()

	cache.Messages = append(cache.Messages, &ConversationMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})

	cleanupExpiredMessages(cache)
}

func GetConversationHistory(channelID string) []ConversationMessage {
	cacheMu.RLock()
	cache, exists := conversationCache[channelID]
	cacheMu.RUnlock()

	if !exists {
		return []ConversationMessage{}
	}

	cache.Mu.RLock()
	defer cache.Mu.RUnlock()

	now := time.Now()
	var history []ConversationMessage

	for _, msg := range cache.Messages {
		if now.Sub(msg.Timestamp) <= cacheExpiry {
			history = append(history, *msg)
		}
	}

	return history
}

func ConvertHistoryToOpenAIMessages(history []ConversationMessage) []openai.ChatCompletionMessageParamUnion {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history))

	for _, msg := range history {
		if msg.Role == "user" {
			messages = append(messages, openai.UserMessage(msg.Content))
		} else if msg.Role == "assistant" {
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	return messages
}

func cleanupExpiredMessages(cache *ChannelCache) {
	if len(cache.Messages) == 0 {
		return
	}

	now := time.Now()
	validMessages := make([]*ConversationMessage, 0)

	for _, msg := range cache.Messages {
		if now.Sub(msg.Timestamp) <= cacheExpiry {
			validMessages = append(validMessages, msg)
		}
	}

	cache.Messages = validMessages
}

func ClearChannelCache(channelID string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	delete(conversationCache, channelID)
}
