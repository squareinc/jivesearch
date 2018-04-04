package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// Speed is an instant answer
type Speed struct {
	Answer
}

func (s *Speed) setQuery(req *http.Request, q string) answerer {
	s.Answer.setQuery(req, q)
	return s
}

func (s *Speed) setUserAgent(req *http.Request) answerer {
	return s
}

func (s *Speed) setLanguage(lang language.Tag) answerer {
	s.language = lang
	return s
}

func (s *Speed) setType() answerer {
	s.Type = "unit converter"
	return s
}

func (s *Speed) setRegex() answerer {
	u := []string{
		"mile",
		"foot", "feet", "ft",
		"kilometer", "km",
		"meter", "knot", "mach",
	}

	for i, uu := range u {
		u[i] = uu + "[s]?"
	}

	rates := []string{"s", "hr", "second", "hour"}

	units := []string{}
	for _, uu := range u {
		for _, r := range rates {
			units = append(units, fmt.Sprintf("%v per %v", uu, r)) // miles per hour
			units = append(units, fmt.Sprintf("%v/%v", uu, r))     // m/h
		}
	}

	units = append(units, "mph", "kmh")

	lll := strings.Join(units, "|")

	t := fmt.Sprintf("[0-9 ]*?%v to [0-9 ]*?%v", lll, lll)

	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)(?P<remainder>.*)$`, t)))
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)(?P<trigger>%s)$`, t)))

	return s
}

func (s *Speed) solve(r *http.Request) answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	s.Solution = "speed"
	return s
}

func (s *Speed) setCache() answerer {
	s.Cache = true
	return s
}

func (s *Speed) tests() []test {
	d := Data{
		Type:      "unit converter",
		Triggered: true,
		Solution:  "speed",
		Cache:     true,
	}

	tests := []test{
		{
			query:    "mph to kmh",
			expected: []Data{d},
		},
		{
			query:    "miles per hour to feet per second",
			expected: []Data{d},
		},
	}

	return tests
}
