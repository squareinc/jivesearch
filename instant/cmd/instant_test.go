package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/contributors"
	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/wikipedia"
	"golang.org/x/text/language"
)

func TestHandler(t *testing.T) {
	for _, c := range []struct {
		query     string
		userAgent string
		want      *instant.Solution
	}{
		{
			query: "january birthstone",
			want: &instant.Solution{
				Type:         "birthstone",
				Triggered:    true,
				Contributors: contributors.Load([]string{"brentadamson"}),
				Text:         "Garnet",
				Cache:        true,
			},
		},
		{
			query:     "user agent",
			userAgent: "firefox",
			want: &instant.Solution{
				Type:         "user agent",
				Triggered:    true,
				Contributors: contributors.Load([]string{"brentadamson"}),
				Text:         "firefox",
				Cache:        false,
			},
		},
		{
			query: "not an instant answer",
			want:  &instant.Solution{},
		},
	} {
		t.Run(c.query, func(t *testing.T) {
			v := url.Values{}
			v.Set("q", c.query)

			r := &http.Request{
				Form:   v,
				Header: make(http.Header),
			}

			r.Header.Set("User-Agent", c.userAgent)

			conf := &cfg{
				&instant.Instant{
					QueryVar:             "q",
					StackOverflowFetcher: &mockStackOverflowFetcher{},
					WikiDataFetcher:      &mockWikiFetcher{},
				},
			}

			rr := httptest.NewRecorder()
			http.HandlerFunc(conf.handler).ServeHTTP(rr, r)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: expected %v got %v",
					http.StatusOK, status)
			}

			got := &instant.Solution{}

			if err := json.NewDecoder(rr.Body).Decode(got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"basic",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := setup()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

type mockStackOverflowFetcher struct{}

func (s *mockStackOverflowFetcher) Fetch(query string, tags []string) (stackoverflow.Response, error) {
	resp := stackoverflow.Response{}
	return resp, nil
}

type mockWikiFetcher struct{}

func (mf *mockWikiFetcher) Fetch(query string, lang language.Tag) (*wikipedia.Item, error) {
	return &wikipedia.Item{}, nil
}

func (mf *mockWikiFetcher) Setup() error {
	return nil
}
