package handler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/S42yt/tuubaa-bot/modules/chatgpt/personality"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func GetAIResponse(userPrompt string) (string, error) {
	return GetAIResponseWithHistory(userPrompt, "")
}

func GetAIResponseWithHistory(userPrompt string, channelID string) (string, error) {
	ulog.Info("[AI] GetAIResponse called with prompt: %s (channel: %s)", userPrompt, channelID)

	token := os.Getenv("GPT_TOKEN")
	if token == "" {
		ulog.Error("[AI] GPT_TOKEN environment variable not set")
		return "", fmt.Errorf("GPT_TOKEN environment variable not set")
	}
	ulog.Debug("[AI] GPT_TOKEN found, length: %d", len(token))

	client := openai.NewClient(option.WithAPIKey(token))
	ulog.Debug("[AI] OpenAI client created")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(personality.SystemPrompt),
	}

	if channelID != "" {
		history := GetConversationHistory(channelID)
		ulog.Info("[AI] Adding %d messages from conversation history", len(history))
		historyMessages := ConvertHistoryToOpenAIMessages(history)
		messages = append(messages, historyMessages...)
	}

	messages = append(messages, openai.UserMessage(userPrompt))

	ulog.Info("[AI] Sending request to OpenAI API with %d messages", len(messages))
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       openai.ChatModelGPT4oMini,
		Messages:    messages,
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(150),
	})
	if err != nil {
		ulog.Error("[AI] OpenAI API error: %v (type: %T)", err, err)
		return "", err
	}
	ulog.Info("[AI] OpenAI API response received")

	if len(resp.Choices) == 0 {
		ulog.Error("[AI] No choices in OpenAI response")
		return "", fmt.Errorf("no response from OpenAI")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	ulog.Info("[AI] Response content: %s", content)
	return content, nil
}

func SendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, response string, userPrompt string) error {
	if len(response) > 2000 {
		response = response[:1997] + "..."
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
}

func SendErrorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, errMsg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Err: %s", errMsg),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
