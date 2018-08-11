package frontend

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jivesearch/jivesearch/bangs"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/log"
	"github.com/jivesearch/jivesearch/search"
	img "github.com/jivesearch/jivesearch/search/image"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

// Context holds a user's request context so we can pass it to our template's form.
// Query, Language, and Region are the RAW query string variables.
type Context struct {
	Q            string          `json:"query"`
	L            string          `json:"-"`
	R            string          `json:"-"`
	N            string          `json:"-"`
	T            string          `json:"-"`
	Safe         bool            `json:"-"`
	DefaultBangs []DefaultBang   `json:"-"`
	Preferred    []language.Tag  `json:"-"`
	Region       language.Region `json:"-"`
	Number       int             `json:"-"`
	Page         int             `json:"-"`
}

// DefaultBang is the user's preffered !bang
type DefaultBang struct {
	Trigger string
	bangs.Bang
}

// Results is the results from search, instant, wikipedia, etc
type Results struct {
	Alternative string          `json:"alternative"`
	Images      *img.Results    `json:"images"`
	Instant     instant.Data    `json:"instant"`
	Search      *search.Results `json:"search"`
}

// Instant is a wrapper to facilitate custom unmarshalling
type Instant struct {
	instant.Data
}

type data struct {
	Brand
	MapBoxKey string
	Context   `json:"-"`
	Results
}

func (f *Frontend) defaultBangs(r *http.Request) []DefaultBang {
	var bngs []DefaultBang

	for _, db := range strings.Split(strings.TrimSpace(r.FormValue("b")), ",") {
		for _, b := range f.Bangs.Bangs {
			for _, t := range b.Triggers {
				if t == db {
					bngs = append(bngs, DefaultBang{db, b})
				}
			}
		}
	}

	if len(bngs) > 0 {
		return bngs
	}

	// defaults if no valid params passed
	for _, b := range []struct {
		trigger string
		name    string
	}{
		{"g", "Google"},
		{"b", "Bing"},
		{"a", "Amazon"},
		{"yt", "YouTube"},
	} {
		for _, bng := range f.Bangs.Bangs {
			if bng.Name == b.name {
				bngs = append(bngs, DefaultBang{b.trigger, bng})
			}
		}
	}

	return bngs
}

// Detect the user's preferred language(s).
// The "l" param takes precedence over the "Accept-Language" header.
func (f *Frontend) detectLanguage(r *http.Request) []language.Tag {
	preferred := []language.Tag{}
	if lang := strings.TrimSpace(r.FormValue("l")); lang != "" {
		if l, err := language.Parse(lang); err == nil {
			preferred = append(preferred, l)
		}
	}

	tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil {
		log.Info.Println(err)
		return preferred
	}

	preferred = append(preferred, tags...)
	return preferred
}

// Detect the user's region. "r" param takes precedence over the language's region (if any).
func (f *Frontend) detectRegion(lang language.Tag, r *http.Request) language.Region {
	reg, err := language.ParseRegion(strings.TrimSpace(r.FormValue("r")))
	if err != nil {
		reg, _ = lang.Region()
	}

	return reg.Canonicalize()
}

func (f *Frontend) addQuery(q string) error {
	exists, err := f.Suggest.Exists(q)
	if err != nil {
		return err
	}

	if !exists {
		if err := f.Suggest.Insert(q); err != nil {
			return err
		}
	}

	return f.Suggest.Increment(q)
}

func (f *Frontend) searchHandler(w http.ResponseWriter, r *http.Request) *response {
	q := strings.TrimSpace(r.FormValue("q"))
	var safe = true
	if strings.TrimSpace(r.FormValue("safe")) == "f" {
		safe = false
	}

	resp := &response{
		status: http.StatusOK,
		data: data{
			Brand:     f.Brand,
			MapBoxKey: f.MapBoxKey,
			Context: Context{
				Safe: safe,
			},
		},
		template: "search",
		err:      nil,
	}

	// render start page if no query
	if q == "" {
		return resp
	}

	d := data{
		f.Brand,
		f.MapBoxKey,
		Context{
			Q:            q,
			L:            strings.TrimSpace(r.FormValue("l")),
			N:            strings.TrimSpace(r.FormValue("n")),
			R:            strings.TrimSpace(r.FormValue("r")),
			T:            strings.TrimSpace(r.FormValue("t")),
			Safe:         safe,
			DefaultBangs: f.defaultBangs(r),
		},
		Results{
			Search: &search.Results{},
		},
	}

	d.Context.Preferred = f.detectLanguage(r)
	lang, _, _ := f.Document.Matcher.Match(d.Context.Preferred...) // will use first supported tag in case of error

	d.Context.Region = f.detectRegion(lang, r)

	// is it a !bang? Redirect them
	if loc, ok := f.Bangs.Detect(d.Context.Q, d.Context.Region, lang); ok {
		return &response{
			status:   302,
			redirect: loc,
		}
	}

	// Let's get them their results
	// what page are they on? Give them first page by default
	var err error
	d.Context.Page, err = strconv.Atoi(strings.TrimSpace(r.FormValue("p")))
	if err != nil || d.Context.Page < 1 {
		d.Context.Page = 1
	}

	// how many results wanted?
	d.Context.Number, err = strconv.Atoi(strings.TrimSpace(r.FormValue("n")))
	if err != nil || d.Context.Number > 100 {
		d.Context.Number = 25
	}

	channels := 1
	imageCH := make(chan *img.Results)
	sc := make(chan *search.Results)
	var ac chan error
	var ic chan instant.Data

	strt := time.Now() // we already have total response time in nginx...we want the breakdown

	if d.Context.Page == 1 {
		channels++
		ac = make(chan error)
		go func(q string, ch chan error) {
			ch <- f.addQuery(q)
		}(d.Context.Q, ac)

		//if d.Context.T != "images" {
		channels++
		ic = make(chan instant.Data)
		go func(r *http.Request) {
			lang, _, _ := f.Wikipedia.Matcher.Match(d.Context.Preferred...)
			key := cacheKey("instant", lang, f.detectRegion(lang, r), r.URL)

			v, err := f.Cache.Get(key)
			if err != nil {
				log.Info.Println(err)
			}

			if v != nil {
				ir := &Instant{
					instant.Data{},
				}

				if err := json.Unmarshal(v.([]byte), &ir); err != nil {
					log.Info.Println(err)
				}

				ic <- ir.Data
				return
			}

			// only need to trigger the maps instant answer if maps or images nav selected
			var onlyMaps bool
			if d.Context.T == "maps" || d.Context.T == "images" {
				onlyMaps = true
			}

			res := f.DetectInstantAnswer(r, lang, onlyMaps)
			if res.Cache {
				var d = f.Cache.Instant

				switch res.Type {
				case "fedex", "ups", "usps", "stock quote", "weather": // only weather with a zip code gets cached "weather 90210"
					d = 1 * time.Minute
				}

				if d > f.Cache.Instant {
					d = f.Cache.Instant
				}

				if err := f.Cache.Put(key, res, d); err != nil {
					log.Info.Println(err)
				}
			}

			ic <- res
		}(r)
		//}
	}

	go func(d data, lang language.Tag, region language.Region) {
		switch d.Context.T {
		case "images":
			key := cacheKey("images", lang, region, r.URL)

			v, err := f.Cache.Get(key)
			if err != nil {
				log.Info.Println(err)
			}

			if v != nil {
				sr := &img.Results{}
				if err := json.Unmarshal(v.([]byte), &sr); err != nil {
					log.Info.Println(err)
				}

				imageCH <- sr
				return
			}

			num := 100
			offset := d.Context.Page*num - num
			sr, err := f.Images.Fetch(d.Context.Q, d.Context.Safe, num, offset) // .8 is Yahoo's open_nsfw cutoff for nsfw
			if err != nil {
				log.Info.Println(err)
			}

			if err := f.Cache.Put(key, sr, f.Cache.Search); err != nil {
				log.Info.Println(err)
			}

			imageCH <- sr
		case "maps":
			resp.template = "maps"
			channels--
		default:
			key := cacheKey("search", lang, region, r.URL)

			v, err := f.Cache.Get(key)
			if err != nil {
				log.Info.Println(err)
			}

			if v != nil {
				sr := &search.Results{}
				if err := json.Unmarshal(v.([]byte), &sr); err != nil {
					log.Info.Println(err)
				}

				sc <- sr
				return
			}

			// get the votes
			offset := d.Context.Page*d.Context.Number - d.Context.Number
			votes, err := f.Vote.Get(d.Context.Q, d.Context.Number*10) // get votes for first 10 pages
			if err != nil {
				log.Info.Println(err)
			}

			sr, err := f.Search.Fetch(d.Context.Q, lang, region, d.Context.Number, offset, votes)
			if err != nil {
				log.Info.Println(err)
			}

			for _, doc := range sr.Documents {
				for _, v := range votes {
					if doc.ID == v.URL {
						doc.Votes = v.Votes
					}
				}
			}

			sr = sr.AddPagination(d.Context.Number, d.Context.Page) // move this to javascript??? (Wouldn't be available in API....)

			if err := f.Cache.Put(key, sr, f.Cache.Search); err != nil {
				log.Info.Println(err)
			}

			sc <- sr
		}

	}(d, lang, d.Context.Region)

	stats := struct {
		autocomplete time.Duration
		images       time.Duration
		instant      time.Duration
		search       time.Duration
	}{}

	for i := 0; i < channels; i++ {
		select {
		case d.Images = <-imageCH:
			// fetch the image & convert to base64 for smoother user experience
			tmp := make(chan *img.Image, len(d.Images.Images))

			go func() {
				for im := range tmp {
					for i, o := range d.Images.Images {
						if im.ID == o.ID {
							d.Images.Images[i] = im
						}
					}
				}
			}()

			var wg sync.WaitGroup

			for _, im := range d.Images.Images {
				wg.Add(1)
				go func(im *img.Image) {
					var err error
					im, err = f.fetchImage(im)
					if err != nil {
						log.Debug.Println(err)
					}
					tmp <- im
					wg.Done()
				}(im)
			}

			wg.Wait()

			stats.images = time.Since(strt).Round(time.Millisecond)
		case d.Instant = <-ic:
			if d.Instant.Err != nil {
				log.Info.Println(d.Instant.Err)
			}
			stats.instant = time.Since(strt).Round(time.Microsecond)
		case d.Search = <-sc:
			stats.search = time.Since(strt).Round(time.Millisecond)
		case err := <-ac:
			if err != nil {
				log.Info.Println(err)
			}
			stats.autocomplete = time.Since(strt).Round(time.Millisecond)
		case <-r.Context().Done():
			// TODO: add info on which items took too long...
			// Perhaps change status code of response so it isn't cached by nginx
			log.Info.Println(errors.Wrapf(r.Context().Err(), "timeout on retrieving results"))
		}
	}

	log.Info.Printf("ac:%v, images: %v, instant (%v):%v, search:%v\n", stats.autocomplete, stats.images, d.Instant.Type, stats.instant, stats.search)

	if r.FormValue("o") == "json" {
		resp.template = r.FormValue("o")
	}

	resp.data = d
	return resp
}

// fetchImage fetches and converts an image to Base64
func (f *Frontend) fetchImage(i *img.Image) (*img.Image, error) {
	var err error

	// go through image proxy to resize and cache the image
	key := hmacKey(i.ID)
	u := fmt.Sprintf("%v/image/225x,s%v/%v", f.Host, key, i.ID)
	fmt.Println(u)

	resp, err := f.Images.Client.Get(u)
	if err != nil {
		return i, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return i, err
	}

	i.Base64 = base64.StdEncoding.EncodeToString(bdy)
	return i, err
}

func cacheKey(item string, lang language.Tag, region language.Region, u *url.URL) string {
	// language and region might be different than what is pass as l & r params
	// ::search::en-US::US::/?q=reverse+%22this%22
	// ::instant::en-US::US::/?q=reverse+%22this%22
	return fmt.Sprintf("::%v::%v::%v::%v", item, lang.String(), region.String(), u.String())
}

// UnmarshalJSON unmarshals an instant answer to the correct data structure
func (d *Instant) UnmarshalJSON(b []byte) error {
	type alias Instant
	raw := &alias{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	j, err := json.Marshal(raw.Solution)
	if err != nil {
		return err
	}

	d.Data = raw.Data
	d.Solution = raw.Solution

	s := detectType(raw.Type)
	if s == nil { // a string
		return nil
	}

	d.Solution = s
	return json.Unmarshal(j, d.Solution)
}

// DetectInstantAnswer triggers the instant answers
func (f *Frontend) DetectInstantAnswer(r *http.Request, lang language.Tag, onlyMaps bool) instant.Data {
	var answers []instant.Answerer

	// select all answers by default, unless user chooses maps
	switch onlyMaps {
	case true:
		answers = []instant.Answerer{
			&instant.Maps{LocationFetcher: f.Instant.LocationFetcher},
			&instant.Wikipedia{
				Fetcher: f.Instant.WikipediaFetcher,
			},
		}
	default:
		answers = []instant.Answerer{
			&instant.BirthStone{},
			&instant.Calculator{},
			&instant.CamelCase{},
			&instant.Characters{},
			&instant.Coin{},
			&instant.Discography{Fetcher: f.Instant.DiscographyFetcher},
			&instant.DigitalStorage{},
			&instant.FedEx{Fetcher: f.Instant.FedExFetcher},
			&instant.Frequency{},
			&instant.FX{Fetcher: f.Instant.FXFetcher},
			&instant.Hash{},
			&instant.Speed{}, // trigger "miles per hour" b/f "miles"
			&instant.Length{},
			&instant.Maps{LocationFetcher: f.Instant.LocationFetcher},
			&instant.Minify{},
			&instant.MortgageCalculator{},
			&instant.Potus{},
			&instant.Power{},
			&instant.Prime{},
			&instant.Random{},
			&instant.Reverse{},
			&instant.Shortener{Service: f.Instant.LinkShortener},
			&instant.Stats{},
			&instant.StockQuote{Fetcher: f.Instant.StockQuoteFetcher},
			&instant.Temperature{},
			&instant.USPS{Fetcher: f.Instant.USPSFetcher},
			&instant.UPS{Fetcher: f.Instant.UPSFetcher},
			&instant.URLDecode{},
			&instant.URLEncode{},
			&instant.UserAgent{},
			&instant.StackOverflow{Fetcher: f.Instant.StackOverflowFetcher},
			&instant.Weather{Fetcher: f.Instant.WeatherFetcher, LocationFetcher: f.Instant.LocationFetcher},
			&instant.Wikipedia{
				Fetcher: f.Instant.WikipediaFetcher,
			}, // always keep this last so that Wikipedia Box will trigger if none other
		}
	}

	// Necessary to use goroutines??? setSolution called only when triggered.
	// Also, the order of some answers matters, like Wikipedia, which is a catch-all
	for _, ia := range answers {
		if triggered := f.Instant.Trigger(ia, r, lang); triggered {
			sol := f.Instant.Solve(ia, r)
			if sol.Err != nil {
				log.Debug.Println(sol.Err)
				continue
			}

			return sol
		}
	}

	return instant.Data{}
}

// detectType returns the proper data structure for an instant answer type
func detectType(t string) interface{} {
	var v interface{}

	switch t {
	case "discography":
		v = &[]discography.Album{}
	case "fx":
		v = &instant.FXResponse{}
	case "fedex", "ups", "usps":
		v = &parcel.Response{}
	case "stackoverflow":
		v = &instant.StackOverflowAnswer{}
	case "stock quote":
		v = &stock.Quote{}
	case "url shortener":
		v = &shortener.Response{}
	case "weather":
		v = &weather.Weather{}
	case "wikipedia":
		v = &wikipedia.Item{}
	case "wikidata age":
		v = &instant.Age{
			Birthday: &instant.Birthday{},
			Death:    &instant.Death{},
		}
	case "wikidata birthday":
		v = &instant.Birthday{}
	case "wikidata death":
		v = &instant.Death{}
	case "wikidata height", "wikidata weight":
		v = &[]wikipedia.Quantity{}
	case "wikiquote":
		v = &[]string{}
	case "wiktionary":
		v = &wikipedia.Wiktionary{}
	default: // a string
		return nil
	}

	return v
}
