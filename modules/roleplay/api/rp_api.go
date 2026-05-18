package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
)

// ----------------//
type APItype int

const (
	OTAKU APItype = iota // 0
	BASTI                // 1
	BOTH                 // 2
)

var API APItype = 0

func GetAPItype() APItype {
	return API
}

func SetAPItype(api APItype) {
	API = api
}

// ----------------//
type gifResponse struct {
	URL string `json:"url"`
}

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

func GetGifURL(kind string) (string, error) {
	cli := &http.Client{Timeout: 8 * time.Second}
	var reqURL string
	var req *http.Request
	var err error
	if API == OTAKU {
		reqURL = fmt.Sprintf("https://api.otakugifs.xyz/gif?reaction=%s", kind)
		req, err = http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return "", err
		}
	} else if API == BASTI {
		key, keyErr := getBastiAPIKey()
		if keyErr != nil {
			return "", keyErr
		}
		reqURL = fmt.Sprintf("https://api.bastiwood.com/reaction/%s", url.PathEscape(kind))
		req, err = http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			API = OTAKU
			return "", err
		}
		applyBastiAuthHeaders(req, key)
	} else if API == BOTH {
		var randAPI int = rand.Intn(2)
		if randAPI == 0 {
			reqURL = fmt.Sprintf("https://api.otakugifs.xyz/gif?reaction=%s", kind)
			req, err = http.NewRequest(http.MethodGet, reqURL, nil)
			if err != nil {
				return "", err
			}
		} else {
			key, keyErr := getBastiAPIKey()
			if keyErr != nil {
				return "", keyErr
			}
			reqURL = fmt.Sprintf("https://api.bastiwood.com/reaction/%s", url.PathEscape(kind))
			req, err = http.NewRequest(http.MethodGet, reqURL, nil)
			if err != nil {
				API = OTAKU
				return "", err
			}
			applyBastiAuthHeaders(req, key)
		}
	}
	ulog.Debug("Fetching GIF for kind=%s url=%s", kind, reqURL)

	resp, err := cli.Do(req)
	if err != nil {
		ulog.Error("GetGifURL: HTTP get failed for kind=%s: %v", kind, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		ulog.Warn("GetGifURL: unexpected status %d for kind=%s", resp.StatusCode, kind)
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var g gifResponse
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		ulog.Error("GetGifURL: json decode failed for kind=%s: %v", kind, err)
		return "", err
	}
	ulog.Debug("GetGifURL: got gif url=%s for kind=%s", g.URL, kind)
	return g.URL, nil
}

func SetGifURL(reaction, gifURL string) error {
	cli := &http.Client{Timeout: 8 * time.Second}
	key, err := getBastiAPIKey()
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"url": gifURL})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.bastiwood.com/setreaction/%s", reaction), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	applyBastiAuthHeaders(req, key)

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
