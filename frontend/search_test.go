package frontend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/bangs"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/search"
	"github.com/jivesearch/jivesearch/search/document"
	"github.com/jivesearch/jivesearch/search/vote"
	"golang.org/x/text/language"
)

func getDefaultBangs() []DefaultBang {
	bngs := bangs.New()

	var m = map[string]bangs.Bang{}

	for _, b := range bngs.Bangs {
		if b.Name == "Google" {
			m["g"] = b
		}
		if b.Name == "Bing" {
			m["b"] = b
		}
		if b.Name == "Amazon" {
			m["a"] = b
		}
		if b.Name == "Youtube" {
			m["yt"] = b
		}
	}

	return []DefaultBang{
		{"g", m["g"]}, {"b", m["b"]}, {"a", m["a"]}, {"yt", m["yt"]},
	}
}

func TestDefaultBangs(t *testing.T) {
	var bing = DefaultBang{}
	var so = DefaultBang{}
	bngs := bangs.New()

	for _, b := range bngs.Bangs {
		if b.Name == "Bing" {
			bing = DefaultBang{"b", b}
		}
		if b.Name == "Stack Overflow" {
			so = DefaultBang{"so", b}
		}
	}

	for _, c := range []struct {
		name string
		args string
		want []DefaultBang
	}{
		{
			"default", "", getDefaultBangs(),
		},
		{
			"custom", "b,so", []DefaultBang{bing, so},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{
				Bangs: bangs.New(),
			}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("b", c.args)
			req.URL.RawQuery = q.Encode()

			got := f.defaultBangs(req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	for _, c := range []struct {
		name           string
		acceptLanguage string
		l              string
		want           []language.Tag
	}{
		{
			"blank", "", "", []language.Tag{},
		},
		{
			"basic", "", "en", []language.Tag{language.English},
		},
		{
			"french", "", "fr", []language.Tag{language.French},
		},
		{
			"Accept-Language header",
			"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7",
			"",
			[]language.Tag{
				language.MustParse("fr-CH"),
				language.French,
				language.English,
				language.German,
			},
		},
		{
			"param overrides Accept-Language header",
			"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7",
			"hr",
			[]language.Tag{
				language.Croatian,
				language.MustParse("fr-CH"),
				language.French,
				language.English,
				language.German,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Accept-Language", c.acceptLanguage)

			q := req.URL.Query()
			q.Add("l", c.l)

			req.URL.RawQuery = q.Encode()

			got := f.detectLanguage(req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestDetectRegion(t *testing.T) {
	for _, c := range []struct {
		name string
		lang language.Tag
		r    string
		want language.Region
	}{
		{
			"empty", language.Tag{}, "", language.MustParseRegion("US").Canonicalize(),
		},
		{
			"basic", language.Tag{}, "us", language.MustParseRegion("US").Canonicalize(),
		},
		{
			"region from language", language.BrazilianPortuguese, "", language.MustParseRegion("BR").Canonicalize(),
		},
		{
			"param overrides language's region", language.CanadianFrench, "gb", language.MustParseRegion("GB").Canonicalize(),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("r", c.r)

			req.URL.RawQuery = q.Encode()

			got := f.detectRegion(c.lang, req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestSearchHandler(t *testing.T) {
	db := getDefaultBangs()

	for _, c := range []struct {
		name     string
		language string
		query    string
		output   string
		want     *response
	}{
		{
			"empty", "en", "", "",
			&response{
				status:   http.StatusOK,
				template: "search",
				data:     nil,
			},
		},
		{
			"basic", "en", " some query ", "",
			&response{
				status:   http.StatusOK,
				template: "search",
				data: data{
					Context: Context{
						Q:            "some query",
						L:            "en",
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
					},
					Results: Results{
						Instant: mockInstantAnswer,
						Search:  mockSearchResults,
					},
				},
			},
		},
		{
			"not cached", "en", "not cached", "",
			&response{
				status:   http.StatusOK,
				template: "search",
				data: data{
					Context: Context{
						Q:            "not cached",
						L:            "en",
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
					},
					Results: Results{
						Instant: mockInstantAnswer,
						Search:  mockSearchResults,
					},
				},
			},
		},
		{
			"json", "en", " some query", "json",
			&response{
				status:   http.StatusOK,
				template: "json",
				data: data{
					Context: Context{
						Q:            "some query",
						L:            "en",
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
					},
					Results: Results{
						Instant: mockInstantAnswer,
						Search:  mockSearchResults,
					},
				},
			},
		},
		{
			"!bang", "", "!g something", "",
			&response{
				status:   http.StatusFound,
				redirect: "https://encrypted.google.com/search?hl=en&q=something",
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			var matcher = language.NewMatcher(
				[]language.Tag{
					language.English,
					language.French,
				},
			)

			f := &Frontend{
				Document: Document{
					Matcher: matcher,
				},
				Bangs: bangs.New(),
				Instant: &instant.Instant{
					WikipediaFetcher:     &mockWikipediaFetcher{},
					StackOverflowFetcher: &mockStackOverflowFetcher{},
				},
				Suggest: &mockSuggester{},
				Search:  &mockSearch{},
				Wikipedia: Wikipedia{
					Matcher: matcher,
				},
				Vote: &mockVoter{},
			}
			f.Cache.Cacher = &mockCacher{}
			f.Cache.Instant = 10 * time.Second
			f.Cache.Search = 10 * time.Second

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("q", c.query)
			q.Add("l", c.language)
			q.Add("o", c.output)
			req.URL.RawQuery = q.Encode()

			got := f.searchHandler(httptest.NewRecorder(), req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestDetectType(t *testing.T) {
	for _, c := range []struct {
		name string
		want interface{}
	}{
		{"birthstone", nil},
		{"fedex", &parcel.Response{}},
		{"stackoverflow", &instant.StackOverflowAnswer{}},
		{"stock quote", &stock.Quote{}},
		{"weather", &weather.Weather{}},
		{"wikipedia", &wikipedia.Item{}},
		{
			"wikidata age", &instant.Age{
				Birthday: &instant.Birthday{},
				Death:    &instant.Death{},
			},
		},
		{"wikidata birthday", &instant.Birthday{}},
		{"wikidata death", &instant.Death{}},
		{"wikidata height", &[]wikipedia.Quantity{}},
		{"wikiquote", &[]string{}},
		{"wiktionary", &wikipedia.Wiktionary{}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := detectType(c.name)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

type mockSearch struct{}

func (s *mockSearch) Fetch(q string, lang language.Tag, region language.Region, page int, number int, votes []vote.Result) (*search.Results, error) {
	r := &search.Results{
		Count:      int64(25),
		Page:       "1",
		Next:       "2",
		Last:       "72",
		Pagination: []string{"2", "3", "4", "5"},
		Documents:  []*document.Document{},
	}

	return r, nil
}

type mockCacher struct{}

func (c *mockCacher) Get(key string) (interface{}, error) {
	var v interface{}

	switch key {
	case "::instant::en::US::/?l=en&o=json&q=+some+query", "::instant::en::US::/?l=en&o=&q=+some+query+":
		v = mockInstantAnswer
	case "::search::en::US::/?l=en&o=json&q=+some+query", "::search::en::US::/?l=en&o=&q=+some+query+":
		v = mockSearchResults
	default:
		return nil, nil
	}

	j, err := json.Marshal(v)
	if err != nil {
		return []int8{}, err
	}

	return j, nil
}

func (c *mockCacher) Put(key string, value interface{}, ttl time.Duration) error {
	return nil
}

// mock Stack Overflow Fetcher
type mockStackOverflowFetcher struct{}

func (s *mockStackOverflowFetcher) Fetch(query string, tags []string) (stackoverflow.Response, error) {
	return stackoverflow.Response{}, nil
}

// mock Wikipedia Fetcher
type mockWikipediaFetcher struct{}

func (mf *mockWikipediaFetcher) Fetch(query string, lang language.Tag) (*wikipedia.Item, error) {
	return &wikipedia.Item{}, nil

}

func (mf *mockWikipediaFetcher) Setup() error {
	return nil
}

var mockInstantAnswer = instant.Data{
	Type:      "wikipedia",
	Triggered: true,
	Solution:  &wikipedia.Item{},
	Err:       nil,
	Cache:     true,
}

var mockSearchResults = &search.Results{
	Count:      int64(25),
	Page:       "1",
	Previous:   "",
	Next:       "2",
	Last:       "72",
	Pagination: []string{"1"},
	Documents:  []*document.Document{},
}
