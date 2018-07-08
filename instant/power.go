package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// Power is an instant answer
type Power struct {
	Answer
}

func (p *Power) setQuery(req *http.Request, q string) Answerer {
	p.Answer.setQuery(req, q)
	return p
}

func (p *Power) setUserAgent(req *http.Request) Answerer {
	return p
}

func (p *Power) setLanguage(lang language.Tag) Answerer {
	p.language = lang
	return p
}

func (p *Power) setType() Answerer {
	p.Type = "unit converter"
	return p
}

func (p *Power) setRegex() Answerer {
	u := []string{
		"watt", "kilowatt", "megawatt", "gigawatt", "terawatt", "petawatt", "exawatt", "horsepower", "hp",
	}

	for i, ll := range u {
		u[i] = fmt.Sprintf(`%v[s]{0,1}\b`, ll)
	}

	lll := strings.Join(u, "|")

	t := fmt.Sprintf(`[0-9]*\s?%v to [0-9]*\s?%v`, lll, lll)

	p.regex = append(p.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)(?P<remainder>.*)$`, t)))
	p.regex = append(p.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)(?P<trigger>%s)$`, t)))

	return p
}

func (p *Power) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	p.Solution = "power"
	return p
}

func (p *Power) setCache() Answerer {
	p.Cache = true
	return p
}

func (p *Power) tests() []test {
	typ := "unit converter"

	d := Data{
		Type:      typ,
		Triggered: true,
		Solution:  "power",
		Cache:     true,
	}

	tests := []test{
		{
			query:    "horsepower to watt",
			expected: []Data{d},
		},
		{
			query:    "megawatt to kilowatt",
			expected: []Data{d},
		},
		{
			query:    "terawatt to hp",
			expected: []Data{d},
		},
	}

	return tests
}
