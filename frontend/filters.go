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
	"Add":              add,
	"Commafy":          commafy,
	"Percent":          percent,
	"SafeHTML":         safeHTML,
	"Truncate":         truncate,
	"HMACKey":          hmacKey,
	"InstantFormatter": instantFormatter,
	"JSONMarshal":      jsonMarshal,
	"Source":           source,
	"Now":              now,
	"WeatherCode":      weatherCode,
	"WikipediaItem":    wikipediaItem,
	"WikiCanonical":    wikiCanonical,
	"WikiDateTime":     wikiDateTime,
	"WikiYears":        wikiYears,
	"WikiLabel":        wikiLabel,
	"WikiJoin":         wikiJoin,
	"WikiAmount":       wikiAmount,
}

func add(x, y int) int {
	return x + y
}

// where did this come from?
func commafy(v interface{}) string {
	switch v.(type) {
	case int64:
		return humanize.Comma(v.(int64))
	case float64:
		return humanize.Commaf(v.(float64))
	default:
		log.Debug.Printf("unknown type %T\n", reflect.TypeOf(v))
		return ""
	}
}

func percent(v float64) string {
	return strconv.FormatFloat(v*100, 'f', 2, 64) + "%"
}

func safeHTML(value string) template.HTML {
	return template.HTML(value)
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

func instantFormatter(sol instant.Data, r language.Region) string {
	switch sol.Solution.(type) {
	case string:
		return sol.Solution.(string)
	case []wikipedia.Quantity: // e.g. height, weight, etc.
		i := sol.Solution.([]wikipedia.Quantity)
		if len(i) == 0 {
			return ""
		}
		return wikiAmount(i[0], r)
	case instant.Age:
		a := sol.Solution.(instant.Age)

		// alive
		if reflect.DeepEqual(a.Death.Death, wikipedia.DateTime{}) {
			return fmt.Sprintf(`<em>Age:</em> %d Years<br><span style="color:#666;">%v</span>`,
				wikiYears(a.Birthday.Birthday, now()), wikiDateTime(a.Birthday.Birthday))
		}

		// dead
		return fmt.Sprintf(`<em>Age at Death:</em> %d Years<br><span style="color:#666;">%v - %v</span>`,
			wikiYears(a.Birthday.Birthday, a.Death.Death), wikiDateTime(a.Birthday.Birthday), wikiDateTime(a.Death.Death))
	case instant.Birthday:
		b := sol.Solution.(instant.Birthday)
		return wikiDateTime(b.Birthday)
	case instant.Death:
		d := sol.Solution.(instant.Death)
		return wikiDateTime(d.Death)
	case wikipedia.Wiktionary: // Wiktionary
		createLink := func(lang, word, style string) string {
			// if this breaks the dump file has the "wiki" key in their json e.g. "enwiktionary", etc.
			return fmt.Sprintf(`<a href="https://%v.wiktionary.org/wiki/%v" %v>%v</a>`, lang, word, style, word)
		}

		def := sol.Solution.(wikipedia.Wiktionary)
		var s = fmt.Sprintf(`<p><span style="font-size:18px;"><em>%v</em></span></p>`, createLink(def.Language, def.Title, `style="color:#333;"`))

		for _, d := range def.Definitions {
			s += fmt.Sprintf(`<span style="font-size:14px;font-style:italic;">%v</span><br>`, d.Part)
			s += fmt.Sprintf(`<span style="display:inline-block;margin-left:15px;">%v</span><br>`, d.Meaning)
			var syn []string
			for _, sy := range d.Synonyms {
				syn = append(syn, createLink(sy.Language, sy.Word, ""))
			}
			if len(syn) > 0 {
				s += fmt.Sprintf(`<span style="display:inline-block;margin-left:15px;font-style:italic;color:#666;">synonyms:&nbsp;</span>%v<br>`,
					strings.Join(syn, ", "),
				)
			}
			s += `<br>`
		}

		return s
	default:
		log.Debug.Printf("unknown instant solution type %T\n", sol.Solution)
		return ""
	}
}

func jsonMarshal(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		log.Debug.Println("error:", err)
	}
	return template.JS(b)
}

// source will show the source of an instant answer if data comes from a 3rd party
func source(answer instant.Data) string {
	var txt string
	var u string
	var img string
	var f string

	switch answer.Type {
	case "fedex":
		txt, u = "FedEx", "https://www.fedex.com"
		img = `<img width="12" height="12" alt="fedex" src="/static/favicons/fedex.ico"/>`
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "stackoverflow":
		// TODO: I wasn't able to get both the User's display name and link to their profile or id.
		// Can select one or the other but not both in their filter.
		user := answer.Solution.(instant.StackOverflowAnswer).Answer.User
		img = `<img width="12" height="12" alt="stackoverflow" src="/static/favicons/stackoverflow.ico"/>`
		f = fmt.Sprintf(`%v via %v <a href="https://stackoverflow.com/">Stack Overflow</a>`, user, img)
	case "stock quote":
		q := answer.Solution.(*stock.Quote)
		switch q.Provider {
		case stock.IEXProvider:
			img = `<img width="12" height="12" alt="iex" src="/static/favicons/iex.ico"/>`
			f = fmt.Sprintf(`%v Data provided for free by <a href="https://iextrading.com/developer">IEX</a>.`, img) // MUST say "Data provided for free by <a href="https://iextrading.com/developer">IEX</a>."
		default:
			log.Debug.Printf("unknown stock quote provider %v\n", q.Provider)
		}
	case "ups":
		txt, u = "UPS", "https://www.ups.com"
		img = `<img width="12" height="12" alt="ups" src="/static/favicons/ups.ico"/>`
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "usps":
		txt, u = "USPS", "https://www.usps.com"
		img = `<img width="12" height="12" alt="usps" src="/static/favicons/usps.ico"/>`
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "weather":
		w := answer.Solution.(*weather.Weather)
		switch w.Provider {
		case weather.OpenWeatherMapProvider:
			txt, u = "OpenWeatherMap", "http://openweathermap.org"
			img = `<img width="12" height="12" alt="openweathermap" src="/static/favicons/openweathermap.ico"/>`
			f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
		default:
			log.Debug.Printf("unknown weather provider %v\n", w.Provider)
		}

	case "wikidata":
		txt, u = "Wikipedia", "https://www.wikipedia.org/"
		img = `<img width="12" height="12" alt="wikipedia" src="/static/favicons/wikipedia.ico"/>`
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "wikiquote":
		txt, u = "Wikiquote", "https://www.wikiquote.org/"
		img = `<img width="12" height="12" alt="wikiquote" src="/static/favicons/wikiquote.ico"/>`
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	case "wiktionary":
		txt, u = "Wiktionary", "https://www.wiktionary.org/"
		img = `<img width="12" height="12" alt="wiktionary" src="/static/favicons/wiktionary.ico"/>`
		f = fmt.Sprintf(`%v <a href="%v">%v</a>`, img, u, txt)
	default:
		log.Debug.Printf("unknown instant answer type %v\n", answer.Type)
	}

	return f
}

var now = func() time.Time { return time.Now().UTC() }

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
		icon = "icon-snow"
	case weather.ThunderStorm:
		icon = "icon-cloud-flash"
	case weather.Windy:
		icon = "icon-windy"
	default:
		icon = "icon-sun"
	}

	return icon
}

func wikipediaItem(sol instant.Data) *wikipedia.Item {
	return sol.Solution.(*wikipedia.Item)
}

// wikiCanonical returns the canonical form of a wikipedia title.
// if this breaks Wikidata dumps have "sitelinks"
func wikiCanonical(t string) string {
	return strings.Replace(t, " ", "_", -1)
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

// wikiJoin joins a slice of Wikidata items
func wikiJoin(items []wikipedia.Wikidata, preferred []language.Tag) string {
	sl := []string{}
	for _, item := range items {
		sl = append(sl, wikiLabel(item.Labels, preferred))
	}

	return strings.Join(sl, ", ")
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
