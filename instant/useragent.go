package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
)

// UserAgent is an instant answer
type UserAgent struct {
	Answer
}

func (u *UserAgent) setQuery(r *http.Request, qv string) answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *UserAgent) setUserAgent(r *http.Request) answerer {
	u.Answer.userAgent = r.UserAgent()
	return u
}

func (u *UserAgent) setType() answerer {
	u.Type = "user agent"
	return u
}

func (u *UserAgent) setContributors() answerer {
	u.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)

	return u
}

func (u *UserAgent) setRegex() answerer {
	triggers := []string{
		"user agent", "user agent?",
		"useragent", "useragent?",
		"my user agent", "my user agent?",
		"my useragent", "my useragent?",
		"what's my user agent", "what's my user agent?",
		"what's my useragent", "what's my useragent?",
		"what is my user agent", "what is my user agent?",
		"what is my useragent", "what is my useragent?",
	}

	t := strings.Join(triggers, "|")
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))

	return u
}

func (u *UserAgent) setSolution() answerer {
	u.Text = u.userAgent

	return u
}

func (u *UserAgent) setCache() answerer {
	// caching would cache the query but the browser could change
	u.Cache = false
	return u
}

func (u *UserAgent) tests() []test {
	typ := "user agent"

	contrib := contributors.Load([]string{"brentadamson"})

	tests := []test{
		{
			query:     "user agent",
			userAgent: "firefox",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "firefox",
					Cache:        false,
				},
			},
		},
		{
			query:     "useragent?",
			userAgent: "opera",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "opera",
					Cache:        false,
				},
			},
		},
		{
			query:     "my user agent",
			userAgent: "some random ua",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "some random ua",
					Cache:        false,
				},
			},
		},
		{
			query:     "what's my user agent?",
			userAgent: "chrome",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "chrome",
					Cache:        false,
				},
			},
		},
		{
			query:     "what is my useragent?",
			userAgent: "internet explorer",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "internet explorer",
					Cache:        false,
				},
			},
		},
	}

	return tests
}
