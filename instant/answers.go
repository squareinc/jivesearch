// Package instant provides instant answers
package instant

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/location"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/log"
	"golang.org/x/text/language"
)

// Instant holds config information for the instant answers
type Instant struct {
	QueryVar             string
	DiscographyFetcher   discography.Fetcher
	FedExFetcher         parcel.Fetcher
	LocationFetcher      location.Fetcher
	StackOverflowFetcher stackoverflow.Fetcher
	StockQuoteFetcher    stock.Fetcher
	UPSFetcher           parcel.Fetcher
	USPSFetcher          parcel.Fetcher
	WeatherFetcher       weather.Fetcher
	WikipediaFetcher     wikipedia.Fetcher
}

// answerer outlines methods for an instant answer
type answerer interface {
	setQuery(r *http.Request, qv string) answerer
	setUserAgent(r *http.Request) answerer
	setLanguage(lang language.Tag) answerer
	setType() answerer
	setRegex() answerer
	trigger() bool
	solve(r *http.Request) answerer
	setCache() answerer
	solution() Data
	tests() []test
}

// Answer holds an instant answer when triggered
type Answer struct {
	query       string
	userAgent   string
	language    language.Tag
	regex       []*regexp.Regexp
	triggerWord string
	remainder   string
	Data
}

// Data holds the returned data of an answer
type Data struct {
	Type      string      `json:"type,omitempty"`
	Triggered bool        `json:"triggered"`
	Solution  interface{} `json:"answer,omitempty"`
	Err       error       `json:"-"`
	Cache     bool        `json:"cache,omitempty"`
}

// Detect loops through all instant answers to find a solution
// Necessary to use goroutines??? setSolution called only when triggered.
func (i *Instant) Detect(r *http.Request, lang language.Tag) Data {
	for _, ia := range i.answers() {
		ia.setUserAgent(r).setQuery(r, i.QueryVar).setLanguage(lang).setRegex()
		if triggered := ia.trigger(); triggered {
			ia.setType().setCache().solve(r)
			return ia.solution()
		}
	}

	return Data{}
}

// setQuery sets the query field
func (a *Answer) setQuery(r *http.Request, qv string) {
	q := strings.ToLower(strings.TrimSpace(r.FormValue(qv)))
	q = strings.Trim(q, "?")
	a.query = strings.Join(strings.Fields(q), " ") // Replace multiple whitespace w/ single whitespace
}

func getIPAddress(r *http.Request) net.IP {
	maxCidrBlocks := []string{
		"127.0.0.1/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"169.254.0.0/16", "::1/128", "fc00::/7", "fe80::/10",
	}

	cidrs := make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}

	isPrivateAddress := func(address string) (bool, error) {
		ipAddress := net.ParseIP(address)
		if ipAddress == nil {
			return false, errors.New("is private address")
		}

		for i := range cidrs {
			if cidrs[i].Contains(ipAddress) {
				return true, nil
			}
		}

		return false, nil
	}

	// Check X-Forward-For and X-Real-IP first
	var ip string
	for _, h := range []string{"X-Forwarded-For", "X-Real-IP"} {
		for _, address := range strings.Split(r.Header.Get(h), ",") {
			ip = strings.TrimSpace(address)
			isPrivate, err := isPrivateAddress(ip)
			if !isPrivate && err == nil {
				return net.ParseIP(ip)
			}
		}
	}

	ip = r.RemoteAddr
	if strings.ContainsRune(r.RemoteAddr, ':') {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return net.ParseIP(ip)
}

// trigger executes the regex for an instant answer
func (a *Answer) trigger() bool {
	for _, re := range a.regex {
		match := re.FindStringSubmatch(a.query)
		if len(match) == 0 {
			continue
		}

		for i, name := range re.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			a.Triggered = true

			switch name {
			case "trigger":
				a.triggerWord = match[i]
			case "remainder":
				a.remainder = match[i]
			default:
				log.Info.Printf("unknown named capture group %q", name)
				return false
			}
		}
		break
	}
	return a.Triggered
}

func (a *Answer) solution() Data {
	return a.Data
}

type test struct {
	query     string
	userAgent string
	ip        net.IP
	expected  []Data
}

// answers returns a slice of all instant answers
// Note: Since we modify fields of the answers we probably shouldn't reuse them....
func (i *Instant) answers() []answerer {
	return []answerer{
		&BirthStone{},
		&Calculator{},
		&CamelCase{},
		&Characters{},
		&Coin{},
		&Discography{Fetcher: i.DiscographyFetcher},
		&DigitalStorage{},
		&FedEx{Fetcher: i.FedExFetcher},
		&Frequency{},
		&Speed{}, // trigger "miles per hour" b/f "miles"
		&Length{},
		&Minify{},
		&Potus{},
		&Power{},
		&Prime{},
		&Random{},
		&Reverse{},
		&Stats{},
		&StockQuote{Fetcher: i.StockQuoteFetcher},
		&Temperature{},
		&USPS{Fetcher: i.USPSFetcher},
		&UPS{Fetcher: i.UPSFetcher},
		&UserAgent{},
		&StackOverflow{Fetcher: i.StackOverflowFetcher},
		&Weather{Fetcher: i.WeatherFetcher, LocationFetcher: i.LocationFetcher},
		&Wikipedia{
			Fetcher: i.WikipediaFetcher,
		}, // always keep this last so that Wikipedia Box will trigger if none other
	}
}

func init() {
	// for coin, random & probably others down the road
	rand.Seed(time.Now().UTC().UnixNano())
}
