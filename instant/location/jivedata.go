package location

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"

	geoip2 "github.com/oschwald/geoip2-golang"
)

// JiveData is a location data provider
type JiveData struct {
	HTTPClient *http.Client
	Key        string
}

// Fetch gets geolocation data from an IP Address
func (j *JiveData) Fetch(ip net.IP) (*geoip2.City, error) {
	u, err := url.Parse("https://jivedata.com")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", j.Key)
	q.Set("ip", ip.String())
	u.RawQuery = q.Encode()

	c := &geoip2.City{}

	resp, err := j.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&c)
	return c, err
}
