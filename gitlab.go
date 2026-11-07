package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dawidl022/gitlab-contribution-sync/contributions"
)

// get all the contributions from the gitlab user via API
func fetchGitlabContributions(username string) (map[contributions.Date]int, error) {
	var contribs map[contributions.Date]int

	host := os.Getenv("GITLAB_HOST") // self-hosted GitLab, e.g. git.example.com
	if host == "" {
		host = "gitlab.com"
	}

	resp, err := http.Get(fmt.Sprintf("https://%s/users/%s/calendar.json", host, username))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&contribs)
	return contribs, err
}
