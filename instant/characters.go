package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// Characters is an instant answer
type Characters struct {
	Answer
}

func (c *Characters) setQuery(r *http.Request, qv string) answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *Characters) setUserAgent(r *http.Request) answerer {
	return c
}

func (c *Characters) setLanguage(lang language.Tag) answerer {
	c.language = lang
	return c
}

func (c *Characters) setType() answerer {
	c.Type = "characters"
	return c
}

func (c *Characters) setRegex() answerer {
	triggers := []string{
		"number of characters in", "number of characters",
		"number of chars in", "number of chars",
		"char count of", "char count",
		"chars count of", "chars count",
		"character count of", "character count",
		"characters count of", "characters count",
		"length in chars", "length in characters",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return c
}

func (c *Characters) solve(r *http.Request) answerer {
	for _, ch := range []string{`"`, `'`} {
		c.remainder = strings.TrimPrefix(c.remainder, ch)
		c.remainder = strings.TrimSuffix(c.remainder, ch)
	}

	c.Solution = strconv.Itoa(len(c.remainder))

	return c
}

func (c *Characters) setCache() answerer {
	c.Cache = true
	return c
}

func (c *Characters) tests() []test {
	typ := "characters"

	tests := []test{
		{
			query: `number of chars in "Jimi Hendrix"`,
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "12",
					Cache:     true,
				},
			},
		},
		{
			query: "number of chars   in Pink   Floyd",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "10",
					Cache:     true,
				},
			},
		},
		{
			query: "Bob Dylan   number of characters in",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "9",
					Cache:     true,
				},
			},
		},
		{
			query: "number of characters Janis   Joplin",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "12",
					Cache:     true,
				},
			},
		},
		{
			query: "char count Led Zeppelin",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "12",
					Cache:     true,
				},
			},
		},
		{
			query: "char count of ' 87 '",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "4",
					Cache:     true,
				},
			},
		},
		{
			query: "they're chars count",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "7",
					Cache:     true,
				},
			},
		},
		{
			query: "chars count of something",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "9",
					Cache:     true,
				},
			},
		},
		{
			query: "Another something chars count of",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "17",
					Cache:     true,
				},
			},
		},
		{
			query: "1234567 character count",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "7",
					Cache:     true,
				},
			},
		},
		{
			query: "character count of house of cards",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "14",
					Cache:     true,
				},
			},
		},
		{
			query: "characters count 50 cent",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "7",
					Cache:     true,
				},
			},
		},
		{
			query: "characters count of 1 dollar",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "8",
					Cache:     true,
				},
			},
		},
		{
			query: "char count equity",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "6",
					Cache:     true,
				},
			},
		},
		{
			query: "characters count seal",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "4",
					Cache:     true,
				},
			},
		},
		{
			query: "length in chars lion",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "4",
					Cache:     true,
				},
			},
		},
		{
			query: "length in characters mountain",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "8",
					Cache:     true,
				},
			},
		},
	}

	return tests
}
