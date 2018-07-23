package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// MortgageCalculator is an instant answer
type MortgageCalculator struct {
	Answer
}

func (c *MortgageCalculator) setQuery(req *http.Request, q string) Answerer {
	c.Answer.setQuery(req, q)
	return c
}

func (c *MortgageCalculator) setUserAgent(req *http.Request) Answerer {
	return c
}

func (c *MortgageCalculator) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *MortgageCalculator) setType() Answerer {
	c.Type = "mortgage calculator"
	return c
}

func (c *MortgageCalculator) setRegex() Answerer {
	t := strings.Join([]string{"mortgage calculator", "calculate mortgage", "mortgage", "mortgage payments"}, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))
	return c
}

func (c *MortgageCalculator) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	return c
}

func (c *MortgageCalculator) setCache() Answerer {
	c.Cache = true
	return c
}

func (c *MortgageCalculator) tests() []test {
	typ := "mortgage calculator"

	d := Data{
		Type:      typ,
		Triggered: true,
		Cache:     true,
	}

	tests := []test{
		{
			query:    "mortgage calculator",
			expected: []Data{d},
		},
	}

	return tests
}
