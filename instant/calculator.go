package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// Calculator is an instant answer
type Calculator struct {
	Answer
}

func (c *Calculator) setQuery(req *http.Request, q string) answerer {
	c.Answer.setQuery(req, q)
	return c
}

func (c *Calculator) setUserAgent(req *http.Request) answerer {
	return c
}

func (c *Calculator) setLanguage(lang language.Tag) answerer {
	c.language = lang
	return c
}

func (c *Calculator) setType() answerer {
	c.Type = "calculator"
	return c
}

func (c *Calculator) setRegex() answerer {
	triggers := []string{
		"calculator", "calculate",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))
	//c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) .*$`, t))) // not implemented yet
	//c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^.* (?P<trigger>%s)$`, t))) // not implemented yet

	return c
}

func (c *Calculator) solve() answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	return c
}

func (c *Calculator) setCache() answerer {
	c.Cache = true
	return c
}

func (c *Calculator) tests() []test {
	typ := "calculator"

	d := Data{
		Type:      typ,
		Triggered: true,
		Cache:     true,
	}

	tests := []test{
		{
			query:    "calculator",
			expected: []Data{d},
		},
		{
			query:    "calculate",
			expected: []Data{d},
		},
	}

	return tests
}
