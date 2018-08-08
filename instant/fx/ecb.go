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
	//rates, err := ecbrates.LoadAll() // full history...takes longer (make method configurable)
	rates, err := ecbrates.Load() // 90 days...sufficient?
	if err != nil {
		return nil, err
	}

	resp := New()

	for _, r := range rates {
		d, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return nil, err
		}

		for c := range r.Rate {
			var currency Currency

			for _, cc := range Currencies {
				if cc.Short == string(c) {
					currency = cc
				}
			}

			if currency == (Currency{}) {
				return nil, err
			}

			rate := &Rate{
				DateTime: d,
			}

			rate.Rate, err = r.Convert(1, c, ecbrates.Currency(resp.Base.Short))
			if err != nil {
				return nil, err
			}

			resp.History[currency] = append(resp.History[currency], rate)
		}
	}

	return resp, err
}
