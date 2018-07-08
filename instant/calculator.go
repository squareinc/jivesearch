package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

// Calculator is an instant answer
type Calculator struct {
	Answer
}

func (c *Calculator) setQuery(req *http.Request, q string) Answerer {
	c.Answer.setQuery(req, q)
	return c
}

func (c *Calculator) setUserAgent(req *http.Request) Answerer {
	return c
}

func (c *Calculator) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *Calculator) setType() Answerer {
	c.Type = "calculator"
	return c
}

var calculatorTriggers = []string{
	"calculator", "calculate", "compute", "formula", "solve", "add", "subtract", "multiply", "divide",
}

func (c *Calculator) setRegex() Answerer {
	t := strings.Join(calculatorTriggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))

	f := `[\s0-9\.\^+\-*\/\(\)]*`
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)?(?P<remainder>%v)$`, t, f)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>%v)(?P<trigger>%s)?$`, f, t)))
	return c
}

func (c *Calculator) solve(r *http.Request) Answerer {
	for _, t := range calculatorTriggers {
		if c.query == t { // eg a search for "calculate", "calculator", etc
			return c
		}

		c.remainder = strings.Replace(c.remainder, t, "", -1)
	}

	if !strings.ContainsAny(c.remainder, "+-/*^") { // don't trigger fedex/ups/usps tracking numbers
		c.Triggered = false
		c.Err = fmt.Errorf("not a mathematical formula %q", c.remainder)
		return c
	}

	expression, err := govaluate.NewEvaluableExpression(c.remainder)
	if err != nil {
		c.Triggered = false
		c.Err = fmt.Errorf("not a mathematical formula %q", c.remainder)
		return c
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		c.Triggered = false
		c.Err = errors.Wrap(err, c.remainder)
		return c
	}

	switch result.(type) {
	case float64:
		c.Solution = result
	default:
		c.Triggered = false
		c.Err = errors.Wrapf(err, "not a mathematical formula %q", c.remainder)
		return c
	}

	return c
}

func (c *Calculator) setCache() Answerer {
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
			query: "calculate 2+2",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  4.0,
					Cache:     true,
				},
			},
		},
		{
			query: "(2+2)*3+6.3",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  18.3,
					Cache:     true,
				},
			},
		},
		{
			query: "(2+2)*3/6.4 compute",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  1.875,
					Cache:     true,
				},
			},
		},
	}

	return tests
}
