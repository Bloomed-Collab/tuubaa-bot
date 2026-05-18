package LLM

import (
	"encoding/json"
	"net/http"
)

const baseURL = "https://api.bastiwood.com"

var (
	prompt string
	loaded bool
)

func loadLLM() error {
	resp, err := http.Post(baseURL+"/loadmodel", "application/json", nil)
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
	resp, err := http.Post(baseURL+"/unloadmodel", "application/json", nil)
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
	url := baseURL + "/generatetext/" + message
	if prompt != "" {
		url += "/" + prompt
	}
	resp, err := http.Get(url) // #nosec G107 — URL is built from a trusted constant base
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Text, nil
}
