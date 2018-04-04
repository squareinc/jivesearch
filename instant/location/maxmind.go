package location

import (
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

// MaxMind is a location data provider
// For install instructions: https://dev.maxmind.com/geoip/geoipupdate/
type MaxMind struct {
	DB string // location of MaxMind database
}

var open = func(loc string) (maxMinder, error) {
	return geoip2.Open(loc)
}

// Fetch gets geolocation data from an IP Address
func (m *MaxMind) Fetch(ip net.IP) (*geoip2.City, error) {
	db, err := open(m.DB)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	c, err := db.City(ip)
	if err != nil {
		return nil, err
	}

	return c, err
}

type maxMinder interface {
	Close() error
	City(ipAddress net.IP) (*geoip2.City, error)
}
