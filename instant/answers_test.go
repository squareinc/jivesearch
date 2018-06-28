package instant

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/location"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	geoip2 "github.com/oschwald/geoip2-golang"
	"golang.org/x/text/language"
)

// TestDetect runs the test cases for each instant answer.
func TestDetect(t *testing.T) {
	cases := []test{}

	i := Instant{
		QueryVar:             "q",
		DiscographyFetcher:   &mockDiscographyFetcher{},
		FedExFetcher:         &mockFedExFetcher{},
		LinkShortener:        &mockShortener{},
		LocationFetcher:      &mockLocationFetcher{},
		StackOverflowFetcher: &mockStackOverflowFetcher{},
		StockQuoteFetcher:    &mockStockQuoteFetcher{},
		UPSFetcher:           &mockUPSFetcher{},
		USPSFetcher:          &mockUSPSFetcher{},
		WeatherFetcher:       &mockWeatherFetcher{},
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
			r.Header.Set("X-Forwarded-For", c.ip.String())

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

func TestGetIPAddress(t *testing.T) {
	type args struct {
		remoteAddr    string
		xRealIP       string
		xForwardedFor []string
	}

	for _, tt := range []struct {
		name string
		args
		want net.IP
	}{
		{
			"no header",
			args{remoteAddr: "161.59.224.138"},
			net.ParseIP("161.59.224.138"),
		},
		{
			"x-real-ip",
			args{xRealIP: "161.59.224.139"},
			net.ParseIP("161.59.224.139"),
		},
		{
			"x-forwarded-for",
			args{xForwardedFor: []string{"161.59.224.140"}},
			net.ParseIP("161.59.224.140"),
		},
		{
			"remote addr and x-forwarded-for",
			args{remoteAddr: "161.59.224.138", xForwardedFor: []string{"161.59.224.140"}},
			net.ParseIP("161.59.224.140"),
		},
		{
			"multiple x-forwarded-for",
			args{xForwardedFor: []string{"161.59.224.140", "161.59.224.141"}},
			net.ParseIP("161.59.224.140"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}

			h.Set("X-Real-IP", tt.args.xRealIP)
			for _, address := range tt.args.xForwardedFor {
				h.Add("X-Forwarded-For", address)
			}

			r := &http.Request{
				RemoteAddr: tt.args.remoteAddr,
				Header:     h,
			}

			got := getIPAddress(r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}

		})
	}
}

// mock FedEx Fetcher
type mockFedExFetcher struct{}

func (f *mockFedExFetcher) Fetch(trackingNumber string) (parcel.Response, error) {
	r := parcel.Response{
		TrackingNumber: strings.ToUpper(trackingNumber),
		Updates: []parcel.Update{
			{
				DateTime: time.Date(2018, 1, 3, 11, 12, 45, 0, time.Local),
				Location: parcel.Location{
					City: "Kandy", State: "ID", Country: "United States",
				},
				Status: "Delivered",
			},
			{
				DateTime: time.Date(2018, 1, 3, 10, 10, 35, 0, time.Local),
				Location: parcel.Location{
					City: "Almost Kandy", State: "ID", Country: "United States",
				},
				Status: "On FedEx vehicle for delivery",
			},
		},
		Expected: parcel.Expected{
			Delivery: "Delivered",
			Date:     time.Date(2018, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		URL: fmt.Sprintf("https://www.fedex.com/apps/fedextrack/?action=track&tracknumbers=%v", strings.ToUpper(trackingNumber)),
	}

	return r, nil
}

// mock location fetcher
type mockLocationFetcher struct{}

func (l *mockLocationFetcher) Fetch(ip net.IP) (*geoip2.City, error) {
	c := &geoip2.City{}
	c.City.Names = map[string]string{"en": "Someville"}
	return c, nil
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

// mock stock quote Fetcher
type mockStockQuoteFetcher struct{}

func (s *mockStockQuoteFetcher) Fetch(ticker string) (*stock.Quote, error) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}

	q := &stock.Quote{}

	switch ticker {
	case "AAPL":
		q.Ticker = "AAPL"
		q.Name = "Apple Inc."
		q.Exchange = stock.NASDAQ
	case "BRK.A":
		q.Ticker = "BRK.A"
		q.Name = "Berkshire Hathaway"
		q.Exchange = stock.NYSE
	}

	q.Last = stock.Last{
		Price:         171.42,
		Time:          time.Unix(1522090355062/1000, 0).In(location),
		Change:        6.48,
		ChangePercent: 0.03929,
	}
	q.History = []stock.EOD{
		{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
		{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
	}
	q.Provider = stock.IEXProvider

	return q, nil
}

// mock UPS Fetcher
type mockUPSFetcher struct{}

func (u *mockUPSFetcher) Fetch(trackingNumber string) (parcel.Response, error) {
	r := parcel.Response{
		TrackingNumber: strings.ToUpper(trackingNumber),
		Updates: []parcel.Update{
			{
				DateTime: time.Date(2018, 3, 11, 2, 38, 0, 0, time.UTC),
				Location: parcel.Location{
					City: "Banahana", State: "ID", Country: "US",
				},
				Status: "Departure Scan",
			},
		},
		Expected: parcel.Expected{
			Delivery: "Scheduled Delivery",
			Date:     time.Date(2018, 3, 11, 0, 0, 0, 0, time.UTC),
		},
		URL: fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", strings.ToUpper(trackingNumber)),
	}

	return r, nil
}

// mock USPS Fetcher
type mockUSPSFetcher struct{}

func (u *mockUSPSFetcher) Fetch(trackingNumber string) (parcel.Response, error) {
	r := parcel.Response{
		TrackingNumber: strings.ToUpper(trackingNumber),
		Updates: []parcel.Update{
			{
				DateTime: time.Date(2018, 3, 12, 13, 57, 0, 0, time.UTC),
				Location: parcel.Location{
					City: "Some City", State: "ID", Country: "",
				},
				Status: "Delivered",
			},
			{
				DateTime: time.Date(2018, 3, 14, 8, 13, 0, 0, time.UTC),
				Location: parcel.Location{
					City: "Close to Some City", State: "ID", Country: "",
				},
				Status: "Out for Delivery",
			},
			{
				DateTime: time.Date(2018, 3, 14, 7, 11, 0, 0, time.UTC),
				Location: parcel.Location{
					City: "Almost", State: "ID", Country: "",
				},
				Status: "Almost there dude",
			},
		},
		URL: fmt.Sprintf("https://tools.usps.com/go/TrackConfirmAction?origTrackNum=%v", strings.ToUpper(trackingNumber)),
	}

	return r, nil
}

// mock weather Fetcher
type mockWeatherFetcher struct {
	location.Fetcher
}

func (m *mockWeatherFetcher) FetchByLatLong(lat, long float64, timezone string) (*weather.Weather, error) {
	w := &weather.Weather{
		City: "Bountiful",
		Current: &weather.Instant{
			Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
			Code:        weather.ScatteredClouds,
			Temperature: 59,
			Low:         55,
			High:        63,
			Wind:        4.7,
			Clouds:      40,
			Rain:        0,
			Snow:        0,
			Pressure:    1014,
			Humidity:    33,
		},
		Forecast: []*weather.Instant{
			{
				Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 97,
				Low:         84,
				High:        97,
				Wind:        3.94,
				Pressure:    888.01,
				Humidity:    14,
			},
			{
				Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 95,
				Low:         85,
				High:        95,
				Wind:        10.76,
				Pressure:    886.87,
				Humidity:    13,
			},
		},
		Provider: weather.OpenWeatherMapProvider,
		TimeZone: "America/Denver",
	}

	return w, nil
}

func (m *mockWeatherFetcher) FetchByZip(zip int) (*weather.Weather, error) {
	w := &weather.Weather{
		City: "Bountiful",
		Current: &weather.Instant{
			Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
			Code:        weather.ScatteredClouds,
			Temperature: 59,
			Low:         55,
			High:        63,
			Wind:        4.7,
			Clouds:      40,
			Rain:        0,
			Snow:        0,
			Pressure:    1014,
			Humidity:    33,
		},
		Forecast: []*weather.Instant{
			{
				Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 97,
				Low:         84,
				High:        97,
				Wind:        3.94,
				Pressure:    888.01,
				Humidity:    14,
			},
			{
				Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 95,
				Low:         85,
				High:        95,
				Wind:        10.76,
				Pressure:    886.87,
				Humidity:    13,
			},
		},
		Provider: weather.OpenWeatherMapProvider,
	}

	return w, nil
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
	default:
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{},
			},
		}, nil
	}
}

func (mf *mockWikipediaFetcher) Setup() error {
	return nil
}

type mockDiscographyFetcher struct{}

func (m *mockDiscographyFetcher) Fetch(artist string) ([]discography.Album, error) {
	u, _ := url.Parse("http://coverartarchive.org/release/1/2-250..jpg")
	return []discography.Album{
		{
			Name:      "Are You Experienced",
			Published: time.Date(1970, 9, 18, 0, 0, 0, 0, time.UTC),
			Image: discography.Image{
				URL: u,
			},
		},
	}, nil
}

type mockShortener struct{}

func (m *mockShortener) Shorten(u *url.URL) (*shortener.Response, error) {
	shrt, _ := url.Parse("http://shrt.url")

	return &shortener.Response{
		Original: u,
		Short:    shrt,
		Provider: "mockShortener",
	}, nil

}
