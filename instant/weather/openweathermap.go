package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// OpenWeatherMap retrieves information from the OpenWeatherMap API
type OpenWeatherMap struct {
	HTTPClient *http.Client
	Key        string
}

// OpenWeatherMapProvider is a weather provider
var OpenWeatherMapProvider provider = "OpenWeatherMap"

type openWeatherMapResponse struct {
	Weather *Weather
	Coord   struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	RawWeather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp     float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity int     `json:"humidity"`
		TempMin  float64 `json:"temp_min"`
		TempMax  float64 `json:"temp_max"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Rain struct {
		All int `json:"3h"` // rain volume for last 3 hrs
	} `json:"rain"`
	Snow struct {
		All int `json:"3h"` // snow volume for last 3 hrs
	} `json:"snow"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int     `json:"type"`
		ID      int     `json:"id"`
		Message float64 `json:"message"`
		Country string  `json:"country"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
	} `json:"sys"`
	ID   int    `json:"id"`
	Name string `json:"name"`
	Cod  int    `json:"cod"`
}

// UnmarshalJSON sets the Response fields
func (r *openWeatherMapResponse) UnmarshalJSON(b []byte) error {
	type alias openWeatherMapResponse
	raw := &alias{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	r.Weather = &Weather{
		City: raw.Name,
		Today: Today{
			Temperature: int(raw.Main.Temp + 0.5),
			Wind:        raw.Wind.Speed,
			Clouds:      float64(raw.Clouds.All),
			Rain:        float64(raw.Rain.All),
			Snow:        float64(raw.Snow.All),
			Pressure:    float64(raw.Main.Pressure),
			Humidity:    float64(raw.Main.Humidity),
			Low:         raw.Main.TempMin,
			High:        raw.Main.TempMax,
		},
	}

	for _, c := range raw.RawWeather {
		if err := r.convertCode(c.ID); err != nil {
			return err
		}
	}

	r.Weather.Provider = OpenWeatherMapProvider

	return err
}

// FetchByLatLong retrieves weather for a latitude/longitude location from the OpenWeatherMap api
func (o *OpenWeatherMap) FetchByLatLong(lat, long float64) (*Weather, error) {
	w := openWeatherMapResponse{}

	u := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?APPID=%v&lat=%v&lon=%v&units=imperial", o.Key, lat, long)

	resp, err := o.HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&w)
	return w.Weather, err
}

// FetchByZip retrieves weather for a zipcode from the OpenWeatherMap api
func (o *OpenWeatherMap) FetchByZip(zip int) (*Weather, error) {
	w := openWeatherMapResponse{}

	u := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?APPID=%v&zip=%d,us&units=imperial", o.Key, zip)

	resp, err := o.HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&w)
	return w.Weather, err
}

func (r *openWeatherMapResponse) convertCode(id int) error {
	// convert OpenWeatherMap code to a weatherCode
	// http://openweathermap.org/weather-conditions
	switch {
	case id >= 200 && id < 300: // thunderstorm w/ light to heavy rain
		r.Weather.Today.Code = ThunderStorm
	case id >= 300 && id < 399: // light to heavy drizzle
		r.Weather.Today.Code = Rain
	case id >= 500 && id < 599: // light to heavy rain
		r.Weather.Today.Code = Rain
	case id >= 600 && id < 699: // light to heavy snow
		r.Weather.Today.Code = Snow
	case id >= 700 && id < 799: // smoke/haze/fog/tornado, etc
		r.Weather.Today.Code = Extreme
	case id == 800: // clear sky
		r.Weather.Today.Code = Clear
	case id == 801:
		r.Weather.Today.Code = LightClouds
	case id == 802, id == 803:
		r.Weather.Today.Code = ScatteredClouds
	case id == 804:
		r.Weather.Today.Code = OvercastClouds
	case id >= 900 && id < 906:
		// extreme tornado/storm/hurricane/cold/heat...we should probably separate extreme cold & heat from others???
		r.Weather.Today.Code = Extreme
	case id >= 951 && id < 999: // low to high wind
		r.Weather.Today.Code = Windy
	default:
		return fmt.Errorf("unknown OpenWeatherMap code %d", id)
	}

	return nil
}
