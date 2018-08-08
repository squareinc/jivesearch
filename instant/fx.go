package instant

import (
	"fmt"
	"net/http"
	"regexp"
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

	resp.Response.Sort()

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
							Base: fx.USD,
							History: map[fx.Currency][]*fx.Rate{
								fx.JPY: {
									{
										DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
										Rate:     1.12,
									},
									{
										DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
										Rate:     1.1,
									},
								},
								fx.GBP: {
									{
										DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
										Rate:     1.5,
									},
									{
										DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
										Rate:     1.6,
									},
								},
							},
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
							Base: fx.USD,
							History: map[fx.Currency][]*fx.Rate{
								fx.JPY: {
									{
										DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
										Rate:     1.12,
									},
									{
										DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
										Rate:     1.1,
									},
								},
								fx.GBP: {
									{
										DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
										Rate:     1.5,
									},
									{
										DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
										Rate:     1.6,
									},
								},
							},
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
