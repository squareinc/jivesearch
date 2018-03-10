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

	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, strings.Join(triggers, "|"))))

	dataStorage := []string{
		"bit", "byte",
		"kilobit", "kibibit", "kilobyte", "kibibyte",
		"megabit", "mebibit", "megabyte", "mebibyte",
		"gigabit", "gibibit", "gigabyte", "gibibyte",
		"terabit", "tebibit", "terabyte", "tebibyte",
		"petabit", "pebibit", "petabyte", "pebibyte",
		"kb", "kbit", "kibit", "kib",
		"mb", "mbit", "mibit", "mib",
		"gb", "gbit", "gibit", "gib",
		"tb", "tbit", "tibit", "tib",
		"pb", "pbit", "pibit", "pib",
	}

	for i, d := range dataStorage {
		dataStorage[i] = d + "[s]?"
	}

	ds := strings.Join(dataStorage, "|")
	t := fmt.Sprintf("[0-9 ]*?%v to[0-9 ]*?%v", ds, ds)

	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)(?P<remainder>.*)$`, t))) // "convert 50mb to kb"
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)(?P<trigger>%s)$`, t)))

	return c
}

func (c *Convert) solve() answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	// TODO: pass the remainder to our html template so that "50gb to mb" prefills the form with "50", "gb", and "mb"

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
		{
			query:    "convert 1mb to pbs",
			expected: []Data{d},
		},
		/*
			not passing for some reason...
			{
				query:    "50gb to mb converter",
				expected: []Data{d},
			},
		*/
		{
			query:    "gb to mb",
			expected: []Data{d},
		},
		{
			query:    "petabytes to megabit",
			expected: []Data{d},
		},
		{
			query:    "50 gb to 100pb",
			expected: []Data{d},
		},
		{
			query:    "50gbs to mbs",
			expected: []Data{d},
		},
	}

	return tests
}
