package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/discography"

	"golang.org/x/text/language"
)

// Discography is an instant answer
type Discography struct {
	discography.Fetcher
	Answer
}

func (d *Discography) setQuery(req *http.Request, q string) answerer {
	d.Answer.setQuery(req, q)
	return d
}

func (d *Discography) setUserAgent(req *http.Request) answerer {
	return d
}

func (d *Discography) setLanguage(lang language.Tag) answerer {
	d.language = lang
	return d
}

func (d *Discography) setType() answerer {
	d.Type = "discography"
	return d
}

func (d *Discography) setRegex() answerer {
	triggers := []string{
		"discography", "albums",
	}

	t := strings.Join(triggers, "|")
	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return d
}

func (d *Discography) solve(r *http.Request) answerer {
	albums, err := d.Fetch(d.remainder)
	if err != nil {
		d.Err = err
		return d
	}

	d.Data.Solution = albums
	return d
}

func (d *Discography) setCache() answerer {
	d.Cache = true
	return d
}

func (d *Discography) tests() []test {
	typ := "discography"

	tests := []test{
		{
			query: "jimi hendrix discography",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: []discography.Album{
						{
							Name:      "Are You Experienced",
							Published: time.Date(1970, 9, 18, 0, 0, 0, 0, time.UTC),
							Image: discography.Image{
								URL: discURL,
							},
						},
					},
					Cache: true,
				},
			},
		},
	}
	return tests
}

var discURL, _ = url.Parse("http://coverartarchive.org/release/1/2-250..jpg")
