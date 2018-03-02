package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

// TestDetect runs the test cases for each instant answer.
func TestDetect(t *testing.T) {
	cases := []test{}

	i := Instant{
		QueryVar:             "q",
		StackOverflowFetcher: &mockStackOverflowFetcher{},
		WikipediaFetcher:     &mockWikipediaFetcher{},
	}

	for j, ia := range i.answers() {
		if len(ia.tests()) == 0 {
			t.Fatalf("No tests for answer #%d", j)
		}
		cases = append(cases, ia.tests()...)
	}

	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			ctx := fmt.Sprintf(`(query: %q, user agent: %q)`, c.query, c.userAgent)

			v := url.Values{}
			v.Set("q", c.query)

			r := &http.Request{
				Form:   v,
				Header: make(http.Header),
			}

			r.Header.Set("User-Agent", c.userAgent)

			got := i.Detect(r, language.English)

			var solved bool

			for _, expected := range c.expected {
				if reflect.DeepEqual(got, expected) {
					solved = true
					break
				}
			}

			if !solved {
				t.Errorf("Instant answer failed %v", ctx)
				t.Errorf("got %+v;", got)
				t.Errorf("want ")
				for _, expected := range c.expected {
					t.Errorf("    %+v\n", expected)
				}
				t.FailNow()
			}
		})
	}
}

// mock Stack Overflow Fetcher
type mockStackOverflowFetcher struct{}

func (s *mockStackOverflowFetcher) Fetch(query string, tags []string) (stackoverflow.Response, error) {
	resp := stackoverflow.Response{}

	switch query {
	case "loop":
		if reflect.DeepEqual(tags, []string{"php"}) {
			resp = stackoverflow.Response{
				Items: []stackoverflow.Item{
					{
						Answers: []stackoverflow.Answer{
							{
								Owner: stackoverflow.Owner{
									DisplayName: "NikiC",
								},
								Score: 1273,
								Body:  "an answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/10057671/how-does-php-foreach-actually-work",
						Title: "How does PHP &#39;foreach&#39; actually work?",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"c++"}) {
			resp = stackoverflow.Response{
				Items: []stackoverflow.Item{
					{
						Answers: []stackoverflow.Answer{
							{
								Owner: stackoverflow.Owner{
									DisplayName: "JamesT",
								},
								Score: 90210,
								Body:  "a very good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/c++-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"go"}) {
			resp = stackoverflow.Response{
				Items: []stackoverflow.Item{
					{
						Answers: []stackoverflow.Answer{
							{
								Owner: stackoverflow.Owner{
									DisplayName: "Danny Zuko",
								},
								Score: 90210,
								Body:  "a superbly good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/go-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"macos"}) {
			resp = stackoverflow.Response{
				Items: []stackoverflow.Item{
					{
						Answers: []stackoverflow.Answer{
							{
								Owner: stackoverflow.Owner{
									DisplayName: "Danny Zuko",
								},
								Score: 90210,
								Body:  "a superbly good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/macos-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"regex"}) {
			resp = stackoverflow.Response{
				Items: []stackoverflow.Item{
					{
						Answers: []stackoverflow.Answer{
							{
								Owner: stackoverflow.Owner{
									DisplayName: "Danny Zuko",
								},
								Score: 90210,
								Body:  "a superbly good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/regex-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		}

	default:
	}

	return resp, nil
}

// mock Wikipedia Fetcher
type mockWikipediaFetcher struct{}

func (mf *mockWikipediaFetcher) Fetch(query string, lang language.Tag) (*wikipedia.Item, error) {
	switch query {
	case "bob marley":
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{
					Birthday: []wikipedia.DateTime{
						{
							Value:    "1945-02-06T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Death: []wikipedia.DateTime{
						{
							Value:    "1981-05-11T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
			},
		}, nil
	case "jimi hendrix":
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{
					Birthday: []wikipedia.DateTime{
						{
							Value:    "1942-11-27T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Death: []wikipedia.DateTime{
						{
							Value:    "1970-09-18T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
			},
		}, nil

	case "shaquille o'neal":
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{
					Height: []wikipedia.Quantity{
						{
							Amount: "2.16",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
					Weight: []wikipedia.Quantity{
						{
							Amount: "147",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
				},
			},
		}, nil
	case "michael jordan":
		return &wikipedia.Item{
			Wikiquote: wikipedia.Wikiquote{
				Quotes: []string{
					"I can accept failure. Everyone fails at something. But I can't accept not trying (no hard work)",
					"ball is life",
				},
			},
		}, nil
	case "guitar":
		return &wikipedia.Item{
			Wiktionary: wikipedia.Wiktionary{
				Title: "guitar",
				Definitions: []*wikipedia.Definition{
					{Part: "noun", Meaning: "musical instrument"},
				},
			},
		}, nil

	}

	return &wikipedia.Item{}, nil

}

func (mf *mockWikipediaFetcher) Setup() error {
	return nil
}
