package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
	"golang.org/x/text/language"
)

// Temperature is an instant answer
type Temperature struct {
	Answer
}

func (t *Temperature) setQuery(r *http.Request, qv string) answerer {
	t.Answer.setQuery(r, qv)
	return t
}

func (t *Temperature) setUserAgent(r *http.Request) answerer {
	return t
}

func (t *Temperature) setLanguage(lang language.Tag) answerer {
	t.language = lang
	return t
}

func (t *Temperature) setType() answerer {
	t.Type = "unit converter"
	return t
}

func (t *Temperature) setContributors() answerer {
	t.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return t
}

func (t *Temperature) setRegex() answerer {
	// a query for "convert" will result in a DigitalStorage answer
	triggers := []string{
		"celsius", "c", "fahrenheit", "f", "temperature converter",
		"temp", "temperature", // when we get weather this should trigger the current weather
	}

	tr := strings.Join(triggers, "|")
	tt := fmt.Sprintf("[0-9 ]*?%v to [0-9 ]*?%v", tr, tr)

	t.regex = append(t.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)(?P<remainder>.*)$`, tt)))
	t.regex = append(t.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)(?P<trigger>%s)$`, tt)))

	return t
}

func (t *Temperature) solve() answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	t.Solution = "temperature"
	return t
}

func (t *Temperature) setCache() answerer {
	t.Cache = true
	return t
}

func (t *Temperature) tests() []test {
	typ := "unit converter"

	contrib := contributors.Load([]string{"brentadamson"})

	d := Data{
		Type:         typ,
		Triggered:    true,
		Contributors: contrib,
		Solution:     "temperature",
		Cache:        true,
	}

	tests := []test{
		{
			query:    "temperature",
			expected: []Data{d},
		},
		{
			query:    "temperature converter",
			expected: []Data{d},
		},
		{
			query:    "17 degrees c to f",
			expected: []Data{d},
		},
		{
			query:    "79.9 f to c",
			expected: []Data{d},
		},
		{
			query:    "107.9 fahrenheit to celsius",
			expected: []Data{d},
		},
		{
			query:    "-9.3 celsius to fahrenheit",
			expected: []Data{d},
		},
	}

	return tests
}
