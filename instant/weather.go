package instant

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	// fetch by lat/long. On localhost this will likely give you weather for "Earth"
	ip := getIPAddress(r)

	city, err := w.LocationFetcher.Fetch(ip)
	if err != nil {
		w.Err = err
	}

	w.Data.Solution, err = w.Fetcher.FetchByLatLong(city.Location.Latitude, city.Location.Longitude, city.Location.TimeZone)
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
						City: "Bountiful",
						Current: &weather.Instant{
							Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Low:         55,
							High:        63,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
						},
						Forecast: []*weather.Instant{
							{
								Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 97,
								Low:         84,
								High:        97,
								Wind:        3.94,
								Pressure:    888.01,
								Humidity:    14,
							},
							{
								Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 95,
								Low:         85,
								High:        95,
								Wind:        10.76,
								Pressure:    886.87,
								Humidity:    13,
							},
						},
						Provider: weather.OpenWeatherMapProvider,
						TimeZone: "America/Denver",
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
						City: "Bountiful",
						Current: &weather.Instant{
							Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Low:         55,
							High:        63,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
						},
						Forecast: []*weather.Instant{
							{
								Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 97,
								Low:         84,
								High:        97,
								Wind:        3.94,
								Pressure:    888.01,
								Humidity:    14,
							},
							{
								Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 95,
								Low:         85,
								High:        95,
								Wind:        10.76,
								Pressure:    886.87,
								Humidity:    13,
							},
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
