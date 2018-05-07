package frontend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/log"
	"golang.org/x/text/language"
)

var funcMap = template.FuncMap{
	"Add":                  add,
	"Commafy":              commafy,
	"Percent":              percent,
	"SafeHTML":             safeHTML,
	"Truncate":             truncate,
	"HMACKey":              hmacKey,
	"Join":                 join,
	"JSONMarshal":          jsonMarshal,
	"Source":               source,
	"Now":                  now,
	"WeatherCode":          weatherCode,
	"WeatherDailyForecast": weatherDailyForecast,
	"WikiAmount":           wikiAmount,
	"WikiCanonical":        wikiCanonical,
	"WikiData":             wikiData,
	"WikiDateTime":         wikiDateTime,
	"WikiJoin":             wikiJoin,
	"WikiLabel":            wikiLabel,
	"WikipediaItem":        wikipediaItem,
	"WikiYears":            wikiYears,
}

func add(x, y int) int {
	return x + y
}

func commafy(v interface{}) string {
	switch v.(type) {
	case int:
		return humanize.Comma(int64(v.(int)))
	case int64:
		return humanize.Comma(v.(int64))
	case float32, float64:
		return humanize.Commaf(v.(float64))
	default:
		log.Debug.Printf("unknown type %T\n", v)
		return ""
	}
}

var hmacSecret = func() string {
	return os.Getenv("hmac_secret")
}

// hmacKey generates an hmac key for our reverse image proxy
func hmacKey(u string) string {
	secret := hmacSecret()
	if secret == "" {
		log.Info.Println(`hmac secret for image proxy is blank. Please set the "hmac_secret" env variable`)
	}

	h := hmac.New(sha256.New, []byte(secret))
	if _, err := h.Write([]byte(u)); err != nil {
		log.Info.Println(err)
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// join joins items in a slice
func join(sl ...string) string {
	var s []string
	for _, item := range sl {
		if item != "" {
			s = append(s, item)
		}
	}

	return strings.Join(s, ", ")
}

func jsonMarshal(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		log.Debug.Println("error:", err)
	}
	return template.JS(b)
}

var now = func() time.Time { return time.Now().UTC() }

func percent(v float64) string {
	return strconv.FormatFloat(v*100, 'f', 2, 64) + "%"
}

func safeHTML(value string) template.HTML {
	return template.HTML(value)
}

// source will show the source of an instant answer if data comes from a 3rd party
func source(answer instant.Data) string {
	var proxyFavIcon = func(u string) string {
		return fmt.Sprintf("/image/32x,s%v/%v", hmacKey(u), u)
	}

	var txt string
	var u string
	var img string
	var f string

	switch answer.Type {
	case "discography":
		txt, u = "MusicBrainz", "https://musicbrainz.org/"
		img = fmt.Sprintf(`<img width="12" height="12" alt="musicbrainz" src="%v"/>`, proxyFavIcon("https://musicbrainz.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "fedex":
		txt, u = "FedEx", "https://www.fedex.com"
		img = fmt.Sprintf(`<img width="12" height="12" alt="fedex" src="%v"/>`, proxyFavIcon("http://www.fedex.com/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "stackoverflow":
		// TODO: I wasn't able to get both the User's display name and link to their profile or id.
		// Can select one or the other but not both in their filter.
		user := answer.Solution.(*instant.StackOverflowAnswer).Answer.User
		img = fmt.Sprintf(`<img width="12" height="12" alt="stackoverflow" src="%v"/>`, proxyFavIcon("https://cdn.sstatic.net/Sites/stackoverflow/img/favicon.ico"))
		f = fmt.Sprintf(`%v via %v <a href="https://stackoverflow.com/">Stack Overflow</a>`, user, img)
	case "stock quote":
		q := answer.Solution.(*stock.Quote)
		switch q.Provider {
		case stock.IEXProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="iex" src="%v"/>`, proxyFavIcon("https://iextrading.com/favicon.ico"))
			f = fmt.Sprintf(`%v Data provided for free by <a href="https://iextrading.com/developer">IEX</a>.`, img) // MUST say "Data provided for free by <a href="https://iextrading.com/developer">IEX</a>."
		default:
			log.Debug.Printf("unknown stock quote provider %v\n", q.Provider)
		}
	case "ups":
		txt, u = "UPS", "https://www.ups.com"
		img = fmt.Sprintf(`<img width="12" height="12" alt="ups" src="%v"/>`, proxyFavIcon("https://www.ups.com/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "usps":
		txt, u = "USPS", "https://www.usps.com"
		img = fmt.Sprintf(`<img width="12" height="12" alt="usps" src="%v"/>`, proxyFavIcon("https://www.usps.com/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "weather":
		w := answer.Solution.(*weather.Weather)
		switch w.Provider {
		case weather.OpenWeatherMapProvider:
			txt, u = "OpenWeatherMap", "http://openweathermap.org"
			img = fmt.Sprintf(`<img width="12" height="12" alt="openweathermap" src="%v"/>`, proxyFavIcon("http://openweathermap.org/favicon.ico"))
			f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
		default:
			log.Debug.Printf("unknown weather provider %v\n", w.Provider)
		}
	case "wikidata age", "wikidata birthday", "wikidata death", "wikidata height", "wikidata weight":
		txt, u = "Wikipedia", "https://www.wikipedia.org/"
		img = fmt.Sprintf(`<img width="12" height="12" alt="wikipedia" src="%v"/>`, proxyFavIcon("https://en.wikipedia.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "wikiquote":
		txt, u = "Wikiquote", "https://www.wikiquote.org/"
		img = fmt.Sprintf(`<img width="12" height="12" alt="wikiquote" src="%v"/>`, proxyFavIcon("https://en.wikiquote.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "wiktionary":
		txt, u = "Wiktionary", "https://www.wiktionary.org/"
		img = fmt.Sprintf(`<img width="12" height="12" alt="wiktionary" src="%v"/>`, proxyFavIcon("https://www.wiktionary.org/static/favicon/piece.ico"))
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	default:
		log.Debug.Printf("unknown instant answer type %v\n", answer.Type)
	}

	return f
}

// Preserving words is a crude translation from the python answer:
// http://stackoverflow.com/questions/250357/truncate-a-string-without-ending-in-the-middle-of-a-word
func truncate(txt string, max int, preserve bool) string {
	if len(txt) <= max {
		return txt
	}

	if preserve {
		c := strings.Fields(txt[:max+1])
		return strings.Join(c[0:len(c)-1], " ") + " ..."
	}

	return txt[:max] + "..."
}

func weatherCode(c weather.Description) string {
	var icon string

	switch c {
	case weather.Clear:
		icon = "icon-sun"
	case weather.LightClouds:
		icon = "icon-cloud-sun"
	case weather.ScatteredClouds:
		icon = "icon-cloud"
	case weather.OvercastClouds:
		icon = "icon-cloud-inv"
	case weather.Extreme:
		icon = "icon-cloud-flash-inv"
	case weather.Rain:
		icon = "icon-rain"
	case weather.Snow:
		icon = "icon-snowflake-o"
	case weather.ThunderStorm:
		icon = "icon-cloud-flash"
	case weather.Windy:
		icon = "icon-windy"
	default:
		icon = "icon-sun"
	}

	return icon
}

type weatherDay struct {
	*weather.Instant
	DT    string
	codes map[weather.Description]int
}

// weatherDailyForecast combines multi-day weather forecasts to 1 daily forecast.
func weatherDailyForecast(forecasts []*weather.Instant, timezone string) []*weatherDay {
	tmp := map[string]*weatherDay{}
	dates := []time.Time{}
	days := []*weatherDay{}

	if timezone == "" { // this is just a hack until we can match timezones with zipcodes
		timezone = "America/Los_Angeles"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		log.Info.Println(err)
	}

	var fmtDate = func(d time.Time) string {
		return d.In(location).Format("Mon 02")
	}

	for _, f := range forecasts {
		fd := fmtDate(f.Date)

		if v, ok := tmp[fd]; ok {
			if f.High > v.Instant.High {
				v.Instant.High = f.High
			}
			if f.Low < v.Instant.Low {
				v.Instant.Low = f.Low
			}
			v.codes[f.Code]++
		} else {
			wd := &weatherDay{
				&weather.Instant{
					Date: f.Date,
					Low:  f.Low,
					High: f.High,
				}, fd, make(map[weather.Description]int),
			}
			wd.codes[f.Code]++
			tmp[fd] = wd
			dates = append(dates, f.Date)
		}
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	for _, d := range dates {
		f := fmtDate(d)
		// find the most frequently used icon for that day
		var most int
		for kk, v := range tmp[f].codes {
			if v > most {
				tmp[f].Code = kk
				most = v
			}
		}

		days = append(days, tmp[f])
	}

	return days
}

// wikiAmount displays a unit in meters, feet, etc depending on user's region
func wikiAmount(q wikipedia.Quantity, r language.Region) string {
	var f string

	amt, err := strconv.ParseFloat(q.Amount, 64)
	if err != nil {
		log.Debug.Println(err)
		return ""
	}

	switch r.String() {
	case "US", "LR", "MM": // only 3 countries that don't use metric system
		switch q.Unit.ID {
		case "Q11573", "Q174728", "Q218593":
			if q.Unit.ID == "Q11573" { // 1 meter = 39.3701 inches
				amt = amt * 39.3701
			} else if q.Unit.ID == "Q174728" { // 1 cm = 0.393701 inches
				amt = amt * .393701
			}

			if amt < 12 {
				f = fmt.Sprintf(`%f"`, amt)
			} else {
				f = fmt.Sprintf(`%d'%d"`, int(amt)/int(12), int(math.Mod(amt, 12)))
			}

		case "Q11570": // 1 kilogram = 2.20462 lbs
			amt = amt * 2.20462
			f = fmt.Sprintf("%d lbs", int(amt+.5))

		default:
			log.Debug.Printf("unknown unit %v\n", q.Unit.ID)
		}
	default:
		s := strconv.FormatFloat(amt, 'f', -1, 64)

		switch q.Unit.ID {
		case "Q11573":
			f = fmt.Sprintf("%v %v", s, "m")
		case "Q174728":
			f = fmt.Sprintf("%v %v", s, "cm")
		case "Q218593":
			amt = amt / .393701
			f = fmt.Sprintf("%v %v", int(amt+.5), "cm")
		case "Q11570":
			f = fmt.Sprintf("%v %v", s, "kg")
		default:
			log.Debug.Printf("unknown unit %v\n", q.Unit.ID)
		}
	}

	return f
}

// wikiCanonical returns the canonical form of a wikipedia title.
// if this breaks Wikidata dumps have "sitelinks"
func wikiCanonical(t string) string {
	return strings.Replace(t, " ", "_", -1)
}

func wikiData(sol instant.Data, r language.Region) string {
	switch sol.Solution.(type) {
	case []wikipedia.Quantity: // height, weight, etc.
		i := sol.Solution.([]wikipedia.Quantity)
		if len(i) == 0 {
			return ""
		}
		return wikiAmount(i[0], r)
	case *[]wikipedia.Quantity: // cached version of height, weight, etc.
		i := *sol.Solution.(*[]wikipedia.Quantity)
		if len(i) == 0 {
			return ""
		}
		return wikiAmount(i[0], r)
	case *instant.Age:
		a := sol.Solution.(*instant.Age)

		// alive
		if a.Death == nil || reflect.DeepEqual(a.Death.Death, wikipedia.DateTime{}) {
			return fmt.Sprintf(`<em>Age:</em> %d Years<br><span style="color:#666;">%v</span>`,
				wikiYears(a.Birthday.Birthday, now()), wikiDateTime(a.Birthday.Birthday))
		}

		// dead
		return fmt.Sprintf(`<em>Age at Death:</em> %d Years<br><span style="color:#666;">%v - %v</span>`,
			wikiYears(a.Birthday.Birthday, a.Death.Death), wikiDateTime(a.Birthday.Birthday), wikiDateTime(a.Death.Death))
	case *instant.Birthday:
		b := sol.Solution.(*instant.Birthday)
		return wikiDateTime(b.Birthday)
	case *instant.Death:
		d := sol.Solution.(*instant.Death)
		return wikiDateTime(d.Death)
	default:
		log.Debug.Printf("unknown instant solution type %T\n", sol.Solution)
		return ""
	}
}

// wikiDateTime formats a date with optional time.
// We assume Gregorian calendar below. (Julian calendar TODO).
// Note: Wikidata only uses Gregorian and Julian calendars.
func wikiDateTime(dt wikipedia.DateTime) string {
	// we loop through the formats until one is found
	// starting with most specific and ending with most general order
	for j, f := range []string{time.RFC3339Nano, "2006"} {
		var ff string

		switch j {
		case 1:
			dt.Value = dt.Value[:4]
			ff = f
		default:
			ff = "January 2, 2006"
		}

		t, err := time.Parse(f, dt.Value)
		if err != nil {
			log.Debug.Println(err)
			continue
		}

		return t.Format(ff)
	}

	return ""
}

func wikipediaItem(sol instant.Data) *wikipedia.Item {
	return sol.Solution.(*wikipedia.Item)
}

// wikiJoin joins a slice of Wikidata items
func wikiJoin(items []wikipedia.Wikidata, preferred []language.Tag) string {
	sl := []string{}
	for _, item := range items {
		sl = append(sl, wikiLabel(item.Labels, preferred))
	}

	return strings.Join(sl, ", ")
}

// wikiLabel extracts the closest label for a Wikipedia Item using a language matcher
func wikiLabel(labels map[string]wikipedia.Text, preferred []language.Tag) string {
	// create a matcher based on the available labels
	langs := []language.Tag{}

	for k := range labels {
		t, err := language.Parse(k)
		if err != nil { // sr-el doesn't parse
			continue
		}

		langs = append(langs, t)
	}

	m := language.NewMatcher(langs)
	lang, _, _ := m.Match(preferred...)

	label := labels[lang.String()]
	return label.Text
}

// wikiYears calculates the number of years (rounded down) betwee two dates.
// e.g. a person's age
func wikiYears(start, end interface{}) int {
	var parseDateTime = func(d interface{}) time.Time {
		switch d.(type) {
		case wikipedia.DateTime:
			dt := d.(wikipedia.DateTime)
			for j, f := range []string{time.RFC3339Nano, "2006"} {
				if j == 1 {
					dt.Value = dt.Value[:4]
				}
				t, err := time.Parse(f, dt.Value)
				if err != nil {
					log.Debug.Println(err)
					continue
				}
				return t
			}

		case time.Time:
			return d.(time.Time)
		default:
			log.Debug.Printf("unknown type %T\n", d)
		}
		return time.Time{}
	}

	s := parseDateTime(start)
	e := parseDateTime(end)

	years := e.Year() - s.Year()
	if e.YearDay() < s.YearDay() {
		years--
	}

	return years
}
