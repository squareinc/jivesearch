package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jivesearch/jivesearch/instant/weather"
	"golang.org/x/text/language"
)

// Weather is an instant answer
type Weather struct {
	Fetcher weather.Fetcher
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
	zip := `[0-9]{5}`

	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)?\s?(?P<remainder>%s)$`, t, zip)))
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>%s)\s(?P<trigger>%s)?$`, zip, t)))

	return w
}

func (w *Weather) solve() answerer {
	z, err := strconv.Atoi(w.remainder)
	if err != nil {
		w.Err = err
		return w
	}

	w.Data.Solution, err = w.Fetcher.FetchByZip(z)
	if err != nil {
		w.Err = err
		return w
	}

	return w
}

func (w *Weather) setCache() answerer {
	w.Cache = true
	return w
}

func (w *Weather) tests() []test {
	typ := "weather"

	tests := []test{
		{
			query: "weather 84014",
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
