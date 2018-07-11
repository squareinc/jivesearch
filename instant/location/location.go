// Package location fetches geolocation data
package location

import (
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

// Location is the lat/long of a place
type Location struct {
	Latitude  float64
	Longitude float64
}

// Fetcher retrieves stock quotes
type Fetcher interface {
	Fetch(ip net.IP) (*geoip2.City, error)
}
