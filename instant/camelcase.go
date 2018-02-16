package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
)

// CamelCase is an instant answer
type CamelCase struct {
	Answer
}

func (c *CamelCase) setQuery(r *http.Request, qv string) answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *CamelCase) setUserAgent(r *http.Request) answerer {
	return c
}

func (c *CamelCase) setType() answerer {
	c.Type = "camelcase"
	return c
}

func (c *CamelCase) setContributors() answerer {
	c.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return c
}

func (c *CamelCase) setRegex() answerer {
	triggers := []string{
		"camelcase",
		"camel case",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return c
}

func (c *CamelCase) setSolution() answerer {
	titled := []string{}
	for _, w := range strings.Fields(c.remainder) {
		titled = append(titled, strings.Title(w))
	}

	c.Text = strings.Join(titled, "")

	return c
}

func (c *CamelCase) setCache() answerer {
	c.Cache = true
	return c
}

func (c *CamelCase) tests() []test {
	typ := "camelcase"

	contrib := contributors.Load([]string{"brentadamson"})

	tests := []test{
		{
			query: "camelcase metallica rocks",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "MetallicaRocks",
					Cache:        true,
				},
			},
		},
		{
			query: "aliCE in chAins Is better camel case",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "AliceInChainsIsBetter",
					Cache:        true,
				},
			},
		},
		{
			query: "camel case O'doyle ruLES",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "O'DoyleRules",
					Cache:        true,
				},
			},
		},
	}

	return tests
}
