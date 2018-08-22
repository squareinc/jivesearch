package instant

import (
	"net/http"
	"regexp"
	"time"

	"github.com/jivesearch/jivesearch/instant/econ/population"
	"github.com/pariz/gountries"
	"golang.org/x/text/language"
)

// Population is an instant answer
type Population struct {
	PopulationFetcher population.Fetcher
	Answer
}

// PopulationResponse is an instant answer response
type PopulationResponse struct {
	Country string
	*population.Response
}

// ErrInvalidCountry indicates a country is not valid
var ErrInvalidCountry error

func (p *Population) setQuery(r *http.Request, qv string) Answerer {
	p.Answer.setQuery(r, qv)
	return p
}

func (p *Population) setUserAgent(r *http.Request) Answerer {
	return p
}

func (p *Population) setLanguage(lang language.Tag) Answerer {
	p.language = lang
	return p
}

func (p *Population) setType() Answerer {
	p.Type = "population"
	return p
}

func (p *Population) setRegex() Answerer {
	p.regex = append(p.regex, regexp.MustCompile(`^(?P<country>.*) population$`))
	p.regex = append(p.regex, regexp.MustCompile(`^(?P<country>.*) population of$`))
	p.regex = append(p.regex, regexp.MustCompile(`^population of (?P<country>.*)$`))
	p.regex = append(p.regex, regexp.MustCompile(`^population (?P<country>.*)$`))

	return p
}

func (p *Population) solve(r *http.Request) Answerer {
	c, ok := p.remainderM["country"]
	if !ok {
		p.Err = ErrInvalidCountry
		return p
	}

	// is it a valid country?
	query := gountries.New()

	country, err := query.FindCountryByName(c)
	if err != nil {
		country, err = query.FindCountryByAlpha(c)
		if err != nil {
			p.Err = err
			return p
		}
	}

	alpha := country.Alpha2

	resp := &PopulationResponse{
		Country: country.Name.Common,
	}

	n := time.Now().Year()
	start := n - 50 // 50 years seems to be the max allowed

	resp.Response, err = p.PopulationFetcher.Fetch(alpha, time.Date(start, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(n, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		p.Err = err
		return p
	}

	resp.Response.Sort()

	p.Data.Solution = resp
	return p
}

func (p *Population) setCache() Answerer {
	p.Cache = true
	return p
}

func (p *Population) tests() []test {
	typ := "population"

	tests := []test{
		{
			query: "Italy population",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &PopulationResponse{
						Country: "IT",
						Response: &population.Response{
							History: []population.Instant{
								{
									Date:  time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC),
									Value: 4,
								},
								{
									Date:  time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC),
									Value: 2,
								},
								{
									Date:  time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC),
									Value: 18,
								},
							},
							Provider: population.TheWorldBankProvider,
						},
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}
