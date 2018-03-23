package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// Reverse is an instant answer
type Reverse struct {
	Answer
}

func (r *Reverse) setQuery(req *http.Request, qv string) answerer {
	r.Answer.setQuery(req, qv)
	return r
}

func (r *Reverse) setUserAgent(req *http.Request) answerer {
	return r
}

func (r *Reverse) setLanguage(lang language.Tag) answerer {
	r.language = lang
	return r
}

func (r *Reverse) setType() answerer {
	r.Type = "reverse"
	return r
}

func (r *Reverse) setRegex() answerer {
	triggers := []string{
		"reverse",
	}

	t := strings.Join(triggers, "|")
	r.regex = append(r.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	r.regex = append(r.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return r
}

func (r *Reverse) solve() answerer {
	for _, c := range []string{`"`, `'`} {
		r.remainder = strings.TrimPrefix(r.remainder, c)
		r.remainder = strings.TrimSuffix(r.remainder, c)
	}

	var n int
	rune := make([]rune, len(r.remainder))
	for _, j := range r.remainder {
		rune[n] = j
		n++
	}
	rune = rune[0:n]

	// Reverse
	for i, j := 0, len(rune)-1; i < j; i, j = i+1, j-1 {
		rune[i], rune[j] = rune[j], rune[i]
	}

	r.Solution = string(rune)

	return r
}

func (r *Reverse) setCache() answerer {
	r.Cache = true
	return r
}

func (r *Reverse) tests() []test {
	typ := "reverse"

	tests := []test{
		{
			query: "reverse ahh lights....ahh see 'em",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "me' ees hha....sthgil hha",
					Cache:     true,
				},
			},
		},
		{
			query: "reverse 私日本語は話せません",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "んせませ話は語本日私",
					Cache:     true,
				},
			},
		},
		{
			query: `reverse "ahh yeah"`,
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution:  "haey hha",
					Cache:     true,
				},
			},
		},
	}

	return tests
}
