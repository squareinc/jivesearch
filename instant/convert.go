package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
	"golang.org/x/text/language"
)

// Convert is an instant answer
type Convert struct {
	Answer
}

func (c *Convert) setQuery(req *http.Request, q string) answerer {
	c.Answer.setQuery(req, q)
	return c
}

func (c *Convert) setUserAgent(req *http.Request) answerer {
	return c
}

func (c *Convert) setLanguage(lang language.Tag) answerer {
	c.language = lang
	return c
}

func (c *Convert) setType() answerer {
	c.Type = "convert"
	return c
}

func (c *Convert) setContributors() answerer {
	c.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return c
}

func (c *Convert) setRegex() answerer {
	triggers := []string{
		"convert", "converter",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))
	//c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) .*$`, t))) // not implemented yet
	//c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^.* (?P<trigger>%s)$`, t))) // not implemented yet

	return c
}

func (c *Convert) solve() answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	return c
}

func (c *Convert) setCache() answerer {
	c.Cache = true
	return c
}

func (c *Convert) tests() []test {
	typ := "convert"

	contrib := contributors.Load([]string{"brentadamson"})

	d := Data{
		Type:         typ,
		Triggered:    true,
		Contributors: contrib,
		Cache:        true,
	}

	tests := []test{
		{
			query:    "convert",
			expected: []Data{d},
		},
	}

	return tests
}
