package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jivesearch/jivesearch/instant/currency"
	"golang.org/x/text/language"
)

// Currency is an instant answer
type Currency struct {
	FXFetcher     currency.FXFetcher
	CryptoFetcher currency.CryptoFetcher
	Answer
}

// CurrencyResponse is an instant answer response
type CurrencyResponse struct {
	*currency.Response
	Notional         float64
	From             currency.Currency
	To               currency.Currency
	ForexCurrencies  []currency.Currency
	CryptoCurrencies []currency.Currency
	Currencies       []currency.Currency
}

// ErrInvalidCurrency indicates the currency was invalid
var ErrInvalidCurrency = fmt.Errorf("invalid currency")

func (c *Currency) setQuery(r *http.Request, qv string) Answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *Currency) setUserAgent(r *http.Request) Answerer {
	return c
}

func (c *Currency) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *Currency) setType() Answerer {
	c.Type = "currency"
	return c
}

func (c *Currency) setRegex() Answerer {
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*)$`))

	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<from>.*)$`))

	c.regex = append(c.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*)$`))

	c.regex = append(c.regex, regexp.MustCompile(`^(?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<from>.*)$`))

	return c
}

func (c *Currency) solve(r *http.Request) Answerer {
	resp := &CurrencyResponse{
		Response:         currency.New(),
		Notional:         1,
		Currencies:       append(currency.ForexCurrencies, currency.CryptoCurrencies...),
		ForexCurrencies:  currency.ForexCurrencies,
		CryptoCurrencies: currency.CryptoCurrencies,
	}

	if n, ok := c.remainderM["notional"]; ok {
		nn, err := strconv.ParseFloat(n, 64)
		if err != nil {
			c.Err = err
			return c
		}
		resp.Notional = nn
	}

	var ok bool

	from := c.remainderM["from"]
	ok, resp.From = currency.Valid(from)
	if !ok {
		c.Err = ErrInvalidCurrency
		return c
	}

	to := c.remainderM["to"]
	ok, resp.To = currency.Valid(to)
	if !ok {
		if resp.From == currency.PHP { // chances are they are looking for a programming answer for PHP
			c.Err = ErrInvalidCurrency
			return c
		}
		resp.To = currency.USD // assume USD for second if not specified "125 BTC"
	}

	cch := make(chan *currency.Response)
	fch := make(chan *currency.Response)
	ech := make(chan error)

	go func(ch chan *currency.Response) {
		crytopResp, err := c.CryptoFetcher.Fetch()
		if err != nil {
			ech <- err
		}
		cch <- crytopResp
	}(cch)

	go func(ch chan *currency.Response) {
		forexResp, err := c.FXFetcher.Fetch()
		if err != nil {
			ech <- err
		}
		fch <- forexResp
	}(fch)

	cresp := &currency.Response{}

	for i := 0; i <= 1; i++ {
		select {
		case res := <-cch:
			cresp.CryptoProvider = res.CryptoProvider
			cresp.History = res.History
		case res := <-fch:
			resp.ForexProvider = res.ForexProvider
			resp.History = res.History
		case err := <-ech:
			c.Err = err
			return c
		}
	}

	resp.CryptoProvider = cresp.CryptoProvider

	// Add crypto to the history
	for k, v := range cresp.History {
		resp.History[k] = v
	}

	resp.Sort()

	c.Data.Solution = resp
	return c
}

func (c *Currency) tests() []test {
	typ := "currency"

	history := map[string][]*currency.Rate{
		currency.JPY.Short: {
			{
				DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
				Rate:     1.12,
			},
			{
				DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
				Rate:     1.1,
			},
		},
		currency.GBP.Short: {
			{
				DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
				Rate:     1.5,
			},
			{
				DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
				Rate:     1.6,
			},
		},
		currency.BTC.Short: {
			{
				DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
				Rate:     1.12,
			},
			{
				DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
				Rate:     1.1,
			},
		},
		currency.LTC.Short: {
			{
				DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
				Rate:     1.5,
			},
			{
				DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
				Rate:     1.6,
			},
		},
	}

	tests := []test{
		{
			query: "convert JPY to USD",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &currency.Response{
							Base:           currency.USD,
							History:        history,
							CryptoProvider: currency.CryptoCompareProvider,
							ForexProvider:  currency.ECBProvider,
						},
						Notional:         1,
						From:             currency.JPY,
						To:               currency.USD,
						Currencies:       append(currency.ForexCurrencies, currency.CryptoCurrencies...),
						ForexCurrencies:  currency.ForexCurrencies,
						CryptoCurrencies: currency.CryptoCurrencies,
					},
				},
			},
		},
		{
			query: "125 EUR to JPY",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &currency.Response{
							Base:           currency.USD,
							History:        history,
							CryptoProvider: currency.CryptoCompareProvider,
							ForexProvider:  currency.ECBProvider,
						},
						Notional:         125,
						From:             currency.EUR,
						To:               currency.JPY,
						Currencies:       append(currency.ForexCurrencies, currency.CryptoCurrencies...),
						ForexCurrencies:  currency.ForexCurrencies,
						CryptoCurrencies: currency.CryptoCurrencies,
					},
				},
			},
		},
		{
			query: "BTC",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &currency.Response{
							Base:           currency.USD,
							History:        history,
							CryptoProvider: currency.CryptoCompareProvider,
							ForexProvider:  currency.ECBProvider,
						},
						Notional:         1,
						From:             currency.BTC,
						To:               currency.USD,
						Currencies:       append(currency.ForexCurrencies, currency.CryptoCurrencies...),
						ForexCurrencies:  currency.ForexCurrencies,
						CryptoCurrencies: currency.CryptoCurrencies,
					},
				},
			},
		},
		{
			query: "125 BTC",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &currency.Response{
							Base:           currency.USD,
							History:        history,
							CryptoProvider: currency.CryptoCompareProvider,
							ForexProvider:  currency.ECBProvider,
						},
						Notional:         125,
						From:             currency.BTC,
						To:               currency.USD,
						Currencies:       append(currency.ForexCurrencies, currency.CryptoCurrencies...),
						ForexCurrencies:  currency.ForexCurrencies,
						CryptoCurrencies: currency.CryptoCurrencies,
					},
				},
			},
		},
	}

	return tests
}
