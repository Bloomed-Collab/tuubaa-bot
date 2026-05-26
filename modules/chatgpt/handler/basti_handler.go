package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// BackendType identifies which LLM provider is active.
type BackendType string

const (
	BackendOpenAI   BackendType = "openai"
	BackendBasti    BackendType = "basti"
	BackendDisabled BackendType = "disabled"
)

// ActiveBackend is the currently active LLM backend. Defaults to Disabled.
var ActiveBackend BackendType = BackendDisabled

var (
	bastiPrompt string
	bastiLoaded bool
)

const bastiBaseURL = "https://api.bastiwood.com"

// IsBastiLoaded reports whether the Basti model is currently loaded.
func IsBastiLoaded() bool { return bastiLoaded }

// SetBastiPrompt sets the system prompt used by the Basti API.
func SetBastiPrompt(p string) { bastiPrompt = p }

func getBastiAPIKey() (string, error) {
	key := os.Getenv("BASTIAPI")
	if key == "" {
		return "", errors.New("BASTIAPI env var is not set")
	}
	return key, nil
}

func applyBastiAuthHeaders(req *http.Request, key string) {
	req.Header.Set("X-API-Key", key)
	req.Header.Set("Authorization", "Bearer "+key)
}

// LoadBastiLLM sends a load-model request to the Basti API.
func LoadBastiLLM() error {
	key, err := getBastiAPIKey()
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, bastiBaseURL+"/loadmodel", nil)
	if err != nil {
		return err
	}
	applyBastiAuthHeaders(req, key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	bastiLoaded = true
	return nil
}

// UnloadBastiLLM sends an unload-model request to the Basti API.
func UnloadBastiLLM() error {
	bastiPrompt = ""
	bastiLoaded = false
	key, err := getBastiAPIKey()
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, bastiBaseURL+"/unloadmodel", nil)
	if err != nil {
		return err
	}
	applyBastiAuthHeaders(req, key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// GetBastiResponse calls the Basti API and returns the generated text.
func GetBastiResponse(message string) (string, error) {
	key, err := getBastiAPIKey()
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("input_text", message)
	if bastiPrompt != "" {
		query.Set("system_prompt", bastiPrompt)
	}
	urlStr := bastiBaseURL + "/generatetext?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	applyBastiAuthHeaders(req, key)
	resp, err := http.DefaultClient.Do(req) // #nosec G107 — URL is built from a trusted constant base
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("basti request failed: status %d: %s", resp.StatusCode, string(bytes.TrimSpace(body)))
	}
	var result struct {
		GeneratedText string `json:"generated_text"`
		Text          string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.GeneratedText != "" {
		return result.GeneratedText, nil
	}
	return result.Text, nil
}
