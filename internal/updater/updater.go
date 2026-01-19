package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func CheckLatestVersion() (string, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/jr-k/d4s/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch release: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}
