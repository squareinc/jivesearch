package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// URLEncode is an instant answer
type URLEncode struct {
	Answer
}

func (u *URLEncode) setQuery(r *http.Request, qv string) Answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *URLEncode) setUserAgent(r *http.Request) Answerer {
	return u
}

func (u *URLEncode) setLanguage(lang language.Tag) Answerer {
	u.language = lang
	return u
}

func (u *URLEncode) setType() Answerer {
	u.Type = "urlencode"
	return u
}

func (u *URLEncode) setRegex() Answerer {
	triggers := []string{
		"urlencode", "encodeurl", "url encode", "encode url", "urlescape", "escapeurl", "url escape", "escape url",
		"uriencode", "encodeuri", "uri encode", "encode uri", "uriescape", "escapeuri", "uri escape", "escape uri",
	}

	t := strings.Join(triggers, "|")
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return u
}

func (u *URLEncode) solve(r *http.Request) Answerer {
	u.Solution = url.QueryEscape(u.remainder)
	return u
}

func (u *URLEncode) setCache() Answerer {
	u.Cache = true
	return u
}

func (u *URLEncode) tests() []test {
	typ := "urlencode"

	tests := []test{
		{
			query: "encode http://www.example.com?q=this|that",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "http%3A%2F%2Fwww.example.com%3Fq%3Dthis%7Cthat",
					Cache:     true,
				},
			},
		},
	}

	return tests
}
