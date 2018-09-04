package congress

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ProPublica holds settings for the ProPublic API
type ProPublica struct {
	Key        string
	HTTPClient *http.Client
}

// ProPublicaProvider is a data Provider
const ProPublicaProvider Provider = "ProPublica"

// FetchSenators returns members of the Congress/Senate from ProPublica
func (p *ProPublica) FetchSenators(location *Location) (*Response, error) {
	u, err := p.buildURL(location)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", p.Key)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ppr := &proPublicaResponse{}
	err = json.NewDecoder(resp.Body).Decode(&ppr)
	if err != nil {
		return nil, err
	}

	r := &Response{
		Location: location,
		Role:     Senators,
		Provider: ProPublicaProvider,
	}
	for _, m := range ppr.Results {
		ne, err := strconv.Atoi(m.NextElection)
		if err != nil {
			return nil, err
		}

		mem := Member{
			Name:         m.Name,
			Gender:       m.Gender,
			Party:        m.Party,
			Twitter:      m.TwitterID,
			Facebook:     m.FacebookAccount,
			NextElection: ne,
		}

		r.Members = append(r.Members, mem)
	}

	return r, err
}

func (p *ProPublica) buildURL(loc *Location) (*url.URL, error) {
	// Senate: https://api.propublica.org/congress/v1/members/{chamber}/{state}/current.json
	// Congress: https://api.propublica.org/congress/v1/members/{chamber}/{state}/{district}/current.json
	return url.Parse(fmt.Sprintf("https://api.propublica.org/congress/v1/members/senate/%v/current.json", loc.Short))
}

type proPublicaResponse struct {
	Status    string `json:"status"`
	Copyright string `json:"copyright"`
	Results   []struct {
		ID              string      `json:"id"`
		Name            string      `json:"name"`
		FirstName       string      `json:"first_name"`
		MiddleName      string      `json:"middle_name"`
		LastName        string      `json:"last_name"`
		Suffix          interface{} `json:"suffix"`
		Role            string      `json:"role"`
		Gender          string      `json:"gender"`
		Party           string      `json:"party"`
		TimesTopicsURL  string      `json:"times_topics_url"`
		TwitterID       string      `json:"twitter_id"`
		FacebookAccount string      `json:"facebook_account"`
		YoutubeID       string      `json:"youtube_id"`
		Seniority       string      `json:"seniority"`
		NextElection    string      `json:"next_election"`
		APIURI          string      `json:"api_uri"`
	} `json:"results"`
}
