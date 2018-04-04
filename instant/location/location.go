// Package location fetches geolocation data
package location

import (
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

// Fetcher retrieves stock quotes
type Fetcher interface {
	Fetch(ip net.IP) (*geoip2.City, error)
}
