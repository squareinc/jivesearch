// Package weather fetches weather data
package weather

// Fetcher retrieves the current and forecasted weather
type Fetcher interface {
	FetchByLatLong(lat, long float64) (*Weather, error)
	FetchByZip(zip int) (*Weather, error)
}

type provider string

// Weather includes the urrent and forecasted weather
type Weather struct {
	City string
	//Updated   time.Time // requires getting the timezone of the user...https://github.com/bradfitz/latlong ???
	Today    Today
	Provider provider
}

// Today is the today's weather
type Today struct {
	Code        Description
	Temperature int
	Wind        float64
	Clouds      float64
	Rain        float64
	Snow        float64
	Pressure    float64
	Humidity    float64
	Low         float64
	High        float64
}

// Description is a standardized weather description
type Description string

// Clear indicates clear skies
const Clear Description = ""

// LightClouds indicate some clouds
const LightClouds Description = "Few Clouds"

// ScatteredClouds indicates scattered clouds
const ScatteredClouds Description = "Scattered Clouds"

// OvercastClouds indicates heavy clouds
const OvercastClouds Description = "Overcast"

// Extreme indicates an extreme event
const Extreme Description = "Extreme"

// Rain indicates rain
const Rain Description = "Rain"

// Snow indicates snow
const Snow Description = "Snow"

// ThunderStorm indicates a thunderstorm
const ThunderStorm Description = "Thunderstorm"

// Windy indicates winds
const Windy Description = "Windy"
