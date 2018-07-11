package instant

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/location"
	"golang.org/x/text/language"
)

// Maps is an instant answer
type Maps struct {
	LocationFetcher location.Fetcher
	Answer
}

func (m *Maps) setQuery(req *http.Request, q string) Answerer {
	m.Answer.setQuery(req, q)
	return m
}

func (m *Maps) setUserAgent(req *http.Request) Answerer {
	return m
}

func (m *Maps) setLanguage(lang language.Tag) Answerer {
	m.language = lang
	return m
}

func (m *Maps) setType() Answerer {
	m.Type = "maps"
	return m
}

func (m *Maps) setRegex() Answerer {
	triggers := []string{
		"map", "maps", "direction", "directions",
	}

	t := strings.Join(triggers, "|")
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) .*$`, t)))
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^.* (?P<trigger>%s)$`, t)))

	return m
}

func (m *Maps) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	ip := getIPAddress(r)

	city, err := m.LocationFetcher.Fetch(ip)
	if err != nil {
		m.Err = err
	}

	m.Data.Solution = location.Location{
		Latitude:  city.Location.Latitude,
		Longitude: city.Location.Longitude,
	}

	return m
}

func (m *Maps) setCache() Answerer {
	m.Cache = true
	return m
}

func (m *Maps) tests() []test {
	typ := "maps"

	tests := []test{
		{
			query: "map",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  location.Location{Latitude: 12, Longitude: 18},
					Cache:     true,
				},
			},
		},
	}

	return tests
}
