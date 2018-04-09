package instant

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jivesearch/jivesearch/instant/location"
	"github.com/jivesearch/jivesearch/instant/weather"
	"golang.org/x/text/language"
)

// Weather is an instant answer
type Weather struct {
	Fetcher         weather.Fetcher
	LocationFetcher location.Fetcher
	Answer
}

func (w *Weather) setQuery(r *http.Request, qv string) answerer {
	w.Answer.setQuery(r, qv)
	return w
}

func (w *Weather) setUserAgent(r *http.Request) answerer {
	return w
}

func (w *Weather) setLanguage(lang language.Tag) answerer {
	w.language = lang
	return w
}

func (w *Weather) setType() answerer {
	w.Type = "weather"
	return w
}

func (w *Weather) setRegex() answerer {
	triggers := []string{
		"climate", "forecast", "weather forecast", "weather",
	}

	t := strings.Join(triggers, "|")

	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)\s?(?P<remainder>.*)$`, t)))
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)\s?(?P<trigger>%s)$`, t)))

	return w
}

func (w *Weather) solve(r *http.Request) answerer {
	if len(w.remainder) == 5 { // US zipcodes for now???
		if z, err := strconv.Atoi(w.remainder); err == nil {
			w.Data.Solution, err = w.Fetcher.FetchByZip(z)
			if err != nil {
				w.Err = err
			}
			w.Cache = true
			return w
		}
	}

	// fetch by lat/long
	// Note: on localhost this will not return anything
	ip := getIPAddress(r)

	city, err := w.LocationFetcher.Fetch(ip)
	if err != nil {
		w.Err = err
	}

	w.Data.Solution, err = w.Fetcher.FetchByLatLong(city.Location.Latitude, city.Location.Longitude)
	if err != nil {
		w.Err = err
	}

	return w
}

func (w *Weather) setCache() answerer {
	// caching is set in solve()
	return w
}

func (w *Weather) tests() []test {
	typ := "weather"

	tests := []test{
		{
			query: "weather",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &weather.Weather{
						City: "Someville",
						Today: weather.Today{
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
							Low:         55.4,
							High:        62.6,
						},
						Provider: weather.OpenWeatherMapProvider,
					},
					Cache: false,
				},
			},
		},
		{
			query: "weather 84014",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &weather.Weather{
						City: "Centerville",
						Today: weather.Today{
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
							Low:         55.4,
							High:        62.6,
						},
						Provider: weather.OpenWeatherMapProvider,
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}
