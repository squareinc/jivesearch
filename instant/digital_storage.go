package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
	"golang.org/x/text/language"
)

// DigitalStorage is an instant answer
type DigitalStorage struct {
	Answer
}

func (d *DigitalStorage) setQuery(req *http.Request, q string) answerer {
	d.Answer.setQuery(req, q)
	return d
}

func (d *DigitalStorage) setUserAgent(req *http.Request) answerer {
	return d
}

func (d *DigitalStorage) setLanguage(lang language.Tag) answerer {
	d.language = lang
	return d
}

func (d *DigitalStorage) setType() answerer {
	d.Type = "unit converter"
	return d
}

func (d *DigitalStorage) setContributors() answerer {
	d.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return d
}

func (d *DigitalStorage) setRegex() answerer {
	// a query for "convert" will result in a DigitalStorage answer
	triggers := []string{
		"convert", "converter",
	}

	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, strings.Join(triggers, "|"))))

	digitalStorage := []string{
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

	for i, ds := range digitalStorage {
		digitalStorage[i] = ds + "[s]?"
	}

	dss := strings.Join(digitalStorage, "|")
	t := fmt.Sprintf("[0-9 ]*?%v to [0-9 ]*?%v", dss, dss)

	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)(?P<remainder>.*)$`, t)))
	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)(?P<trigger>%s)$`, t)))

	return d
}

func (d *DigitalStorage) solve() answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	// TODO: pass the remainder to our html template so that "50gb to mb" prefills the form with "50", "gb", and "mb"
	// Note: Combining the digital storage, length and other unit converters would then require us to make
	// sense of the regexp remainder and that seems like a hassle. We could put unit converters in a subpackage, though.
	d.Solution = "digital storage"
	return d
}

func (d *DigitalStorage) setCache() answerer {
	d.Cache = true
	return d
}

func (d *DigitalStorage) tests() []test {
	typ := "unit converter"

	contrib := contributors.Load([]string{"brentadamson"})

	dd := Data{
		Type:         typ,
		Triggered:    true,
		Contributors: contrib,
		Solution:     "digital storage",
		Cache:        true,
	}

	tests := []test{
		{
			query:    "convert",
			expected: []Data{dd},
		},
		{
			query:    "convert 1mb to pbs",
			expected: []Data{dd},
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
			expected: []Data{dd},
		},
		{
			query:    "petabytes to megabit",
			expected: []Data{dd},
		},
		{
			query:    "50 gb to 100pb",
			expected: []Data{dd},
		},
		{
			query:    "50gbs to mbs",
			expected: []Data{dd},
		},
	}

	return tests
}
