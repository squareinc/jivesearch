package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
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

func (c *CamelCase) setLanguage(lang language.Tag) answerer {
	c.language = lang
	return c
}

func (c *CamelCase) setType() answerer {
	c.Type = "camelcase"
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

func (c *CamelCase) solve(r *http.Request) answerer {
	titled := []string{}
	for _, w := range strings.Fields(c.remainder) {
		titled = append(titled, strings.Title(w))
	}

	c.Solution = strings.Join(titled, "")

	return c
}

func (c *CamelCase) setCache() answerer {
	c.Cache = true
	return c
}

func (c *CamelCase) tests() []test {
	typ := "camelcase"

	tests := []test{
		{
			query: "camelcase metallica rocks",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "MetallicaRocks",
					Cache:     true,
				},
			},
		},
		{
			query: "aliCE in chAins Is better camel case",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "AliceInChainsIsBetter",
					Cache:     true,
				},
			},
		},
		{
			query: "camel case O'doyle ruLES",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "O'DoyleRules",
					Cache:     true,
				},
			},
		},
	}

	return tests
}
