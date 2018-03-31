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

func (p *Power) setQuery(req *http.Request, q string) answerer {
	p.Answer.setQuery(req, q)
	return p
}

func (p *Power) setUserAgent(req *http.Request) answerer {
	return p
}

func (p *Power) setLanguage(lang language.Tag) answerer {
	p.language = lang
	return p
}

func (p *Power) setType() answerer {
	p.Type = "unit converter"
	return p
}

func (p *Power) setRegex() answerer {
	u := []string{
		"mile", "yard", "foot", "feet", "inch", "nautical mile",
		"ft", "in",
		// https://en.wikipedia.org/wiki/Metre (use the common units that are in boldface)
		"centimeter", "millimeter", "micrometer", "nanometer", "meter", "kilometer",
		"centimetre", "millimetre", "micrometre", "nanometre", "metre", "kilometre",
		"cm", "mm", "nm", "km", //"m", // triggers other instant answers and causes tests to fail
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

func (p *Power) solve() answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	l.Solution = "length"
	return l
}

func (p *Power) setCache() answerer {
	l.Cache = true
	return l
}

func (p *Power) tests() []test {
	typ := "unit converter"

	d := Data{
		Type:      typ,
		Triggered: true,
		Solution:  "length",
		Cache:     true,
	}

	tests := []test{
		{
			query:    "ins to cms",
			expected: []Data{d},
		},
		{
			query:    "convert 1 meter to feet",
			expected: []Data{d},
		},
		{
			query:    "10 meters to foot",
			expected: []Data{d},
		},
		{
			query:    "cm to m",
			expected: []Data{d},
		},
		{
			query:    "inches to cm",
			expected: []Data{d},
		},
		{
			query:    "nm to feets", // why not
			expected: []Data{d},
		},
	}

	return tests
}
