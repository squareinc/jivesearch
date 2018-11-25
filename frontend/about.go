package frontend

import (
	"encoding/json"
	"net/http"
)

// GitHub holds settings for GitHub's API
type GitHub struct {
	HTTPClient *http.Client
}

type contributor struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
	Contributions     int    `json:"contributions"`
}

type about struct {
	Brand
	Context      `json:"-"`
	Contributors []*contributor
	Onion        string
}

func (f *Frontend) aboutHandler(w http.ResponseWriter, r *http.Request) *response {
	resp := &response{
		status:   http.StatusOK,
		template: "about",
		data: about{
			Brand: f.Brand,
			Onion: f.Onion,
		},
	}

	/*
		For more detail on additions, deletions and commits: https://api.github.com/repos/jivesearch/jivesearch/stats/contributors
		The below is sorted by # of contributions (in descending order).
	*/
	rsp, err := f.GitHub.HTTPClient.Get("https://api.github.com/repos/jivesearch/jivesearch/contributors")
	if err != nil {
		resp.err = err
		return resp
	}

	defer rsp.Body.Close()

	cont := []*contributor{}
	err = json.NewDecoder(rsp.Body).Decode(&cont)
	if err != nil {
		resp.err = err
		return resp
	}

	resp.data = about{
		Brand:        f.Brand,
		Contributors: cont,
		Onion:        f.Onion,
	}

	return resp
}
