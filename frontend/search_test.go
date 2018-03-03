package frontend

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jivesearch/jivesearch/bangs"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/contributors"
	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/search"
	"github.com/jivesearch/jivesearch/search/document"
	"github.com/jivesearch/jivesearch/search/vote"
	"golang.org/x/text/language"
)

func TestDefaultBang(t *testing.T) {
	var google bangs.Bang
	var bing bangs.Bang

	bngs := bangs.New()

	for _, b := range bngs.Bangs {
		if b.Name == "Google" {
			google = b
		} else if b.Name == "Bing" {
			bing = b
		}
	}

	for k, v := range map[string]bangs.Bang{"Google": google, "Bing": bing} {
		var empty bangs.Bang
		if reflect.DeepEqual(v, empty) {
			t.Fatalf("could not find !bang for %q", k)
		}
	}

	for _, c := range []struct {
		name string
		bang string
		args string
		want DefaultBang
	}{
		{
			"default", "Google", "", DefaultBang{"g", google},
		},
		{
			"bing", "Bing", "b", DefaultBang{"b", bing},
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

			got := f.defaultBang(req)

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
	// get the default !bang
	var bng bangs.Bang
	bngs := bangs.New()

	for _, b := range bngs.Bangs {
		if b.Name == "Google" {
			bng = b
		}
	}

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
						Q:           "some query",
						L:           "en",
						DefaultBang: DefaultBang{"g", bng},
						Preferred:   []language.Tag{language.MustParse("en")},
						Region:      language.MustParseRegion("US"),
						Number:      25,
						Page:        1,
					},
					Results: Results{
						Instant: instant.Data{
							Type:      "wikipedia",
							Triggered: true,
							Contributors: []contributors.Contributor{
								{Name: "Brent Adamson", Github: "brentadamson", Twitter: "thebrentadamson"},
							},
							Solution: &wikipedia.Item{},
							Err:      nil,
							Cache:    true,
						},
						Search: &search.Results{
							Count:      int64(25),
							Page:       "1",
							Previous:   "",
							Next:       "2",
							Last:       "72",
							Pagination: []string{"1"},
							Documents:  []*document.Document{},
						},
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
						Q:           "some query",
						L:           "en",
						DefaultBang: DefaultBang{"g", bng},
						Preferred:   []language.Tag{language.MustParse("en")},
						Region:      language.MustParseRegion("US"),
						Number:      25,
						Page:        1,
					},
					Results: Results{
						Instant: instant.Data{
							Type:      "wikipedia",
							Triggered: true,
							Contributors: []contributors.Contributor{
								{Name: "Brent Adamson", Github: "brentadamson", Twitter: "thebrentadamson"},
							},
							Solution: &wikipedia.Item{},
							Err:      nil,
							Cache:    true,
						},
						Search: &search.Results{
							Count:      int64(25),
							Page:       "1",
							Previous:   "",
							Next:       "2",
							Last:       "72",
							Pagination: []string{"1"},
							Documents:  []*document.Document{},
						},
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
