package frontend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/log"
	"github.com/jivesearch/jivesearch/wikipedia"
	"golang.org/x/text/language"
)

var funcMap = template.FuncMap{
	"Commafy":          commafy,
	"SafeHTML":         safeHTML,
	"Truncate":         truncate,
	"HMACKey":          hmacKey,
	"InstantFormatter": instantFormatter,
	"Now":              now,
	"WikiCanonical":    wikiCanonical,
	"WikiDateTime":     wikiDateTime,
	"WikiYears":        wikiYears,
	"WikiLabel":        wikiLabel,
	"WikiJoin":         wikiJoin,
	"WikiAmount":       wikiAmount,
}

// where did this come from?
func commafy(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = 0 - v
	}
	parts := []string{"", "", "", "", "", "", ""}
	j := len(parts) - 1

	for v > 999 {
		parts[j] = strconv.FormatInt(v%1000, 10)

		switch len(parts[j]) {
		case 2:
			parts[j] = "0" + parts[j]
		case 1:
			parts[j] = "00" + parts[j]
		}
		v = v / 1000
		j--
	}
	parts[j] = strconv.Itoa(int(v))
	return sign + strings.Join(parts[j:], ",")
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
	secret := os.Getenv("hmac_secret")
	if secret == "" {
		log.Info.Println(`hmac secret for image proxy is blank. Please set the "hmac_secret" env variable`)
	}
	return secret
}

// hmacKey generates an hmac key for our reverse image proxy
func hmacKey(u string) string {
	secret := hmacSecret()

	h := hmac.New(sha256.New, []byte(secret))
	if _, err := h.Write([]byte(u)); err != nil {
		log.Info.Println(err)
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func instantFormatter(raw interface{}, r language.Region) string {
	switch raw.(type) {
	case []wikipedia.Quantity: // e.g. height, weight, etc.
		i := raw.([]wikipedia.Quantity)
		if len(i) == 0 {
			return ""
		}
		return wikiAmount(i[0], r)
	case instant.Age:
		a := raw.(instant.Age)
		if a.Alive {
			return fmt.Sprintf(`<em>Age:</em> %d Years<br><span style="color:#666;">%v</span>`,
				wikiYears(a.Birthday.Birthday, time.Now()), wikiDateTime(a.Birthday.Birthday))
		}
		return fmt.Sprintf(`<em>Age at Death:</em> %d Years<br><span style="color:#666;">%v - %v</span>`,
			wikiYears(a.Birthday.Birthday, a.Death.Death), wikiDateTime(a.Birthday.Birthday), wikiDateTime(a.Death.Death))
	case instant.Birthday:
		b := raw.(instant.Birthday)
		return wikiDateTime(b.Birthday)
	case instant.Death:
		d := raw.(instant.Death)
		return fmt.Sprintf("<em>Date of Death:</em> %v", wikiDateTime(d.Death))
	}

	return ""
}

func now() time.Time {
	return time.Now()
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

	log.Info.Println(s, e, s.YearDay(), e.YearDay(), years)

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
