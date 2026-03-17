package handler

import (
	"errors"
	"strings"
	"unicode/utf8"
)

func ValidateAndCleanPrompt(prompt string) (string, error) {
	cleaned := strings.TrimSpace(prompt)

	if cleaned == "" {
		return "", errors.New("prompt cannot be empty")
	}

	if utf8.RuneCountInString(cleaned) > 100 {
		return "", errors.New("prompt cannot exceed 100 characters")
	}

	return cleaned, nil
}

func PreparePromptForAPI(userPrompt string) string {
	return strings.TrimSpace(userPrompt)
}
