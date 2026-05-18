package LLM

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const baseURL = "https://api.bastiwood.com"

var (
	prompt string
	loaded bool
)

func getBastiAPIKey() (string, error) {
	key := os.Getenv("BASTIAPI")
	if key == "" {
		return "", errors.New("BASTIAPI is not set")
	}
	return key, nil
}

func applyBastiAuthHeaders(req *http.Request, key string) {
	req.Header.Set("X-API-Key", key)
	req.Header.Set("Authorization", "Bearer "+key)
}

func loadLLM() error {
	key, err := getBastiAPIKey()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/loadmodel", nil)
	if err != nil {
		return err
	}
	applyBastiAuthHeaders(req, key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	loaded = true
	return nil
}

func unloadLLM() error {
	prompt = ""
	loaded = false
	key, err := getBastiAPIKey()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/unloadmodel", nil)
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

func setprompt(p string) {
	prompt = p
}

func getmessage(message string) (string, error) {
	key, err := getBastiAPIKey()
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("input_text", message)
	if prompt != "" {
		query.Set("system_prompt", prompt)
	}
	urlStr := baseURL + "/generatetext?" + query.Encode()

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
		return "", fmt.Errorf("llm request failed: status %d: %s", resp.StatusCode, string(bytes.TrimSpace(body)))
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
