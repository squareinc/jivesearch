package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
)

var reFrequency *regexp.Regexp

// Frequency is an instant answer
type Frequency struct {
	Answer
}

func (f *Frequency) setQuery(r *http.Request, qv string) answerer {
	f.Answer.setQuery(r, qv)
	return f
}

func (f *Frequency) setUserAgent(r *http.Request) answerer {
	return f
}

func (f *Frequency) setType() answerer {
	f.Type = "frequency"
	return f
}

func (f *Frequency) setContributors() answerer {
	f.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return f
}

func (f *Frequency) setRegex() answerer {
	triggers := []string{
		"frequency of",
	}

	t := strings.Join(triggers, "|")
	f.regex = append(f.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	//f.regex = append(f.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t))) // not implemented yet
	return f
}

func (f *Frequency) setSolution() answerer {
	var char string
	var wrd string

	matches := reFrequency.FindStringSubmatch(f.remainder)
	if len(matches) == 3 {
		char = matches[1]
		wrd = matches[2]
	}

	if char != "" && wrd != "" {
		cnt := 0
		for _, c := range wrd {
			if string(c) == char {
				cnt++
			}
		}
		f.Text = strconv.Itoa(cnt)
	}

	return f
}

func (f *Frequency) setCache() answerer {
	f.Cache = true
	return f
}

func (f *Frequency) tests() []test {
	typ := "frequency"

	contrib := contributors.Load([]string{"brentadamson"})

	tests := []test{
		test{
			query: "a in abracadabra frequency of",
			expected: []Solution{
				Solution{},
			},
		},
		test{
			query: "frequency of",
			expected: []Solution{
				Solution{},
			},
		},
		test{
			query: "frequency of a in abracadabra",
			expected: []Solution{
				Solution{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "5",
					Cache:        true,
				},
			},
		},
		test{
			query: "frequency of o in cooler",
			expected: []Solution{
				Solution{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "2",
					Cache:        true,
				},
			},
		},
		test{
			query: "frequency of s in jimi hendrix",
			expected: []Solution{
				Solution{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "0",
					Cache:        true,
				},
			},
		},
		test{
			query: "frequency of e in fred astaire",
			expected: []Solution{
				Solution{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Text:         "2",
					Cache:        true,
				},
			},
		},
	}

	return tests

}

func init() {
	reFrequency = regexp.MustCompile(`^(.*?) in (.+)`)
}
