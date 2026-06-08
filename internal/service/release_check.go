package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type githubLatestRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type ReleaseCheckResult struct {
	CurrentVersion   string
	LatestRelease    string
	LatestReleaseURL string
	UpdateAvailable  bool
}

func CheckLatestRelease(currentVersion string) (ReleaseCheckResult, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/openinfraops/powershell-dsc-pullserver/releases/latest", nil)
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "go-dsc-pull-release-check")

	resp, err := client.Do(req)
	if err != nil {
		return ReleaseCheckResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ReleaseCheckResult{}, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rel githubLatestRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return ReleaseCheckResult{}, err
	}

	result := ReleaseCheckResult{
		CurrentVersion:   currentVersion,
		LatestRelease:    rel.TagName,
		LatestReleaseURL: rel.HTMLURL,
		UpdateAvailable:  rel.TagName != "" && rel.TagName != currentVersion,
	}
	return result, nil
}

func StartReleaseCheckWorker(currentVersion string, interval time.Duration, onResult func(ReleaseCheckResult, error)) {
	run := func() {
		result, err := CheckLatestRelease(currentVersion)
		if onResult != nil {
			onResult(result, err)
		}
	}

	run()
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			run()
		}
	}()
}
