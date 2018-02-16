package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
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

func (c *Characters) setType() answerer {
	c.Type = "characters"
	return c
}

func (c *Characters) setContributors() answerer {
	c.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
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

func (c *Characters) setSolution() answerer {
	for _, ch := range []string{`"`, `'`} {
		c.remainder = strings.TrimPrefix(c.remainder, ch)
		c.remainder = strings.TrimSuffix(c.remainder, ch)
	}

	c.Text = strconv.Itoa(len(c.remainder))

	return c
}

func (c *Characters) setCache() answerer {
	c.Cache = true
	return c
}

func (c *Characters) tests() []test {
	typ := "characters"

	contrib := contributors.Load([]string{"brentadamson"})

	tests := []test{
		{
			query: `number of chars in "Jimi Hendrix"`,
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "12",
					Cache:        true,
				},
			},
		},
		{
			query: "number of chars   in Pink   Floyd",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "10",
					Cache:        true,
				},
			},
		},
		{
			query: "Bob Dylan   number of characters in",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "9",
					Cache:        true,
				},
			},
		},
		{
			query: "number of characters Janis   Joplin",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "12",
					Cache:        true,
				},
			},
		},
		{
			query: "char count Led Zeppelin",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "12",
					Cache:        true,
				},
			},
		},
		{
			query: "char count of ' 87 '",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "4",
					Cache:        true,
				},
			},
		},
		{
			query: "they're chars count",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "7",
					Cache:        true,
				},
			},
		},
		{
			query: "chars count of something",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "9",
					Cache:        true,
				},
			},
		},
		{
			query: "Another something chars count of",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "17",
					Cache:        true,
				},
			},
		},
		{
			query: "1234567 character count",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "7",
					Cache:        true,
				},
			},
		},
		{
			query: "character count of house of cards",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "14",
					Cache:        true,
				},
			},
		},
		{
			query: "characters count 50 cent",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "7",
					Cache:        true,
				},
			},
		},
		{
			query: "characters count of 1 dollar",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "8",
					Cache:        true,
				},
			},
		},
		{
			query: "chars in saved by the bell",
			expected: []Solution{
				{},
			},
		},
		{
			query: "chars 21 jump street",
			expected: []Solution{
				{},
			},
		},
		{
			query: "characters in house of cards",
			expected: []Solution{
				{},
			},
		},
		{
			query: "characters beavis and butthead",
			expected: []Solution{
				{},
			},
		},
		{
			query: "char count equity",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "6",
					Cache:        true,
				},
			},
		},
		{
			query: "characters count seal",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "4",
					Cache:        true,
				},
			},
		},
		{
			query: "length in chars lion",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "4",
					Cache:        true,
				},
			},
		},
		{
			query: "length in characters mountain",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "8",
					Cache:        true,
				},
			},
		},
		{
			query: "length of 1 meter",
			expected: []Solution{
				{},
			},
		},
	}

	return tests
}
