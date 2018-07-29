package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/fx"
	"golang.org/x/text/language"
)

// FX is an instant answer
type FX struct {
	Fetcher fx.Fetcher
	Answer
}

// FXResponse is an instant answer response
type FXResponse struct {
	*fx.Response
	Notional float64
	From     fx.Currency
	To       fx.Currency
}

// ErrInvalidCurrency indicates the currency was invalid
var ErrInvalidCurrency = fmt.Errorf("invalid currency")

func (f *FX) setQuery(r *http.Request, qv string) Answerer {
	f.Answer.setQuery(r, qv)
	return f
}

func (f *FX) setUserAgent(r *http.Request) Answerer {
	return f
}

func (f *FX) setLanguage(lang language.Tag) Answerer {
	f.language = lang
	return f
}

func (f *FX) setType() Answerer {
	f.Type = "fx"
	return f
}

func (f *FX) setRegex() Answerer {
	f.regex = append(f.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*) to (?P<to>.*)$`))
	f.regex = append(f.regex, regexp.MustCompile(`^convert (?P<from>.*) to (?P<to>.*)$`))
	f.regex = append(f.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*) to (?P<to>.*)$`))
	f.regex = append(f.regex, regexp.MustCompile(`^(?P<from>.*) to (?P<to>.*)$`))
	return f
}

func (f *FX) solve(r *http.Request) Answerer {
	resp := &FXResponse{
		Notional: 1,
	}

	if n, ok := f.remainderM["notional"]; ok {
		nn, err := strconv.ParseFloat(n, 64)
		if err != nil {
			f.Err = err
			return f
		}
		resp.Notional = nn
	}

	// make sure valid currencies were passed in
	for _, c := range fx.Currencies {
		if c.Short == strings.ToUpper(f.remainderM["from"]) {
			resp.From = c
		}
	}

	for _, c := range fx.Currencies {
		if c.Short == strings.ToUpper(f.remainderM["to"]) {
			resp.To = c
		}
	}

	if resp.From == (fx.Currency{}) {
		f.Err = ErrInvalidCurrency
		return f
	}

	if resp.To == (fx.Currency{}) {
		f.Err = ErrInvalidCurrency
		return f
	}

	var err error
	resp.Response, err = f.Fetcher.Fetch()
	if err != nil {
		f.Err = err
		return f
	}

	sort.Slice(resp.Rates, func(i, j int) bool {
		return resp.Rates[i].Currency.Long < resp.Rates[j].Currency.Long
	})

	f.Data.Solution = resp
	return f
}

func (f *FX) setCache() Answerer {
	f.Cache = true
	return f
}

func (f *FX) tests() []test {
	typ := "fx"

	tests := []test{
		{
			query: "convert JPY to USD",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &FXResponse{
						Response: &fx.Response{
							Rates: []*fx.Rate{
								{
									Base:     fx.USD,
									Currency: fx.BGN,
									Rate:     .5944,
								},
								{
									Base:     fx.USD,
									Currency: fx.EUR,
									Rate:     1.1625,
								},
								{
									Base:     fx.USD,
									Currency: fx.JPY,
									Rate:     0.009,
								},
								{
									Base:     fx.USD,
									Currency: fx.USD,
									Rate:     1.0,
								},
							},
							DateTime: time.Date(2018, 07, 27, 0, 0, 0, 0, time.UTC),
							Provider: fx.ECBProvider,
						},
						Notional: 1,
						From:     fx.JPY,
						To:       fx.USD,
					},
					Cache: true,
				},
			},
		},
		{
			query: "125 EUR to JPY",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &FXResponse{
						Response: &fx.Response{
							Rates: []*fx.Rate{
								{
									Base:     fx.USD,
									Currency: fx.BGN,
									Rate:     .5944,
								},
								{
									Base:     fx.USD,
									Currency: fx.EUR,
									Rate:     1.1625,
								},
								{
									Base:     fx.USD,
									Currency: fx.JPY,
									Rate:     0.009,
								},
								{
									Base:     fx.USD,
									Currency: fx.USD,
									Rate:     1.0,
								},
							},
							DateTime: time.Date(2018, 07, 27, 0, 0, 0, 0, time.UTC),
							Provider: fx.ECBProvider,
						},
						Notional: 125,
						From:     fx.EUR,
						To:       fx.JPY,
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}
