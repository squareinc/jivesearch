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
	Code        weatherCode
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

type weatherCode string

// Clear indicates clear skies
const Clear weatherCode = ""

// LightClouds indicate some clouds
const LightClouds weatherCode = "Few Clouds"

// ScatteredClouds indicates scattered clouds
const ScatteredClouds weatherCode = "Scattered Clouds"

// OvercastClouds indicates heavy clouds
const OvercastClouds weatherCode = "Overcast"

// Extreme indicates an extreme event
const Extreme weatherCode = "Extreme"

// Rain indicates rain
const Rain weatherCode = "Rain"

// Snow indicates snow
const Snow weatherCode = "Snow"

// ThunderStorm indicates a thunderstorm
const ThunderStorm weatherCode = "Thunderstorm"

// Windy indicates winds
const Windy weatherCode = "Windy"
