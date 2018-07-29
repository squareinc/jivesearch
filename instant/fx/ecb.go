package fx

import (
	"time"

	"github.com/openprovider/ecbrates"
)

// ECBProvider is a fx provider
var ECBProvider provider = "European Central Bank"

// ECB holds settings for ECB via openprovider
type ECB struct{}

// Fetch retrieves fx quotes from the ECB via openprovider
func (e *ECB) Fetch() (*Response, error) {
	r, err := ecbrates.New()
	if err != nil {
		return nil, err
	}

	d, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		Provider: ECBProvider,
		DateTime: d,
	}

	// convert them to the base rate
	for c := range r.Rate {
		rate := &Rate{}
		rate.setBase()

		rate.Rate, err = r.Convert(1, c, ecbrates.Currency(rate.Base.Short))
		if err != nil {
			return nil, err
		}

		var found bool
		for _, cc := range Currencies {
			if cc.Short == string(c) {
				rate.Currency = cc
				found = true
			}
		}

		if !found {
			return nil, err
		}

		resp.Rates = append(resp.Rates, rate)
	}

	return resp, err
}
