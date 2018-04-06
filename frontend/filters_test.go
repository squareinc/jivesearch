package frontend

import (
	"encoding/json"
	"html/template"
	"reflect"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

func TestAdd(t *testing.T) {
	type args struct {
		x int
		y int
	}

	for _, tt := range []struct {
		name string
		args
		want int
	}{
		{
			name: "1+1",
			args: args{1, 1},
			want: 2,
		},
		{
			name: "103+873",
			args: args{103, 873},
			want: 976,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := add(tt.args.x, tt.args.y)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCommafy(t *testing.T) {
	for _, tt := range []struct {
		number interface{}
		want   string
	}{
		{
			number: 1000,
			want:   "1,000",
		},
		{
			number: 1023,
			want:   "1,023",
		},
		{
			number: -12000000,
			want:   "-12,000,000",
		},
		{
			number: -120,
			want:   "-120",
		},
		{
			number: 999000.01,
			want:   "999,000.01",
		},
		{
			number: 48915619813218,
			want:   "48,915,619,813,218",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := commafy(tt.number)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestHMACKey(t *testing.T) {
	type args struct {
		u      string
		secret string
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{"http://www.example.com/some/path/?query=string", "my_secret"},
			want: "LGSCFXg045ByB4ShdCHRIDlrPUDJ9eyFSrGz0HrtfAo=",
		},
		{
			name: "empty secret",
			args: args{"http://www.example.com/some/path/?query=string", ""},
			want: "oz13AtRiNq7h_rBVZMXccxPnDfnVHR12zd4honudDk4=",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hmacSecret = func() string { return tt.args.secret }

			got := hmacKey(tt.args.u)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestJSONMarshall(t *testing.T) {
	for _, tt := range []struct {
		name string
		arg  interface{}
	}{
		{
			"json object", `{"name":"bob"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.arg)
			if err != nil {
				t.Fatal(err)
			}
			want := template.JS(b)

			got := jsonMarshal(tt.arg)
			if got != want {
				t.Fatalf("got %q; want %q", got, want)
			}
		})
	}
}

func TestJoinLocation(t *testing.T) {
	for _, tt := range []struct {
		sl   []string
		want string
	}{
		{
			sl:   []string{"salt lake city", "utah", ""},
			want: "salt lake city, utah",
		},
		{
			sl:   []string{"boston", "MA", "US"},
			want: "boston, MA, US",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := joinLocation(tt.sl[0], tt.sl[1], tt.sl[2])
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	for _, tt := range []struct {
		number float64
		want   string
	}{
		{
			number: .2357,
			want:   "23.57%",
		},
		{
			number: .01527,
			want:   "1.53%",
		},
		{
			number: -1.0,
			want:   "-100.00%",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := percent(tt.number)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSafeHTML(t *testing.T) {
	for _, tt := range []struct {
		arg  string
		want template.HTML
	}{
		{
			arg:  "<!--[if lte IE 8]>",
			want: "<!--[if lte IE 8]>",
		},
		{
			arg:  "<!--[if gt IE 8]><!-->",
			want: "<!--[if gt IE 8]><!-->",
		},
		{
			arg:  "<script>some nasty javascript</script>",
			want: "<script>some nasty javascript</script>",
		},
	} {
		t.Run(tt.arg, func(t *testing.T) {
			got := safeHTML(tt.arg)

			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSource(t *testing.T) {
	type args struct {
		src instant.Data
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "empty",
			args: args{instant.Data{}},
			want: "",
		},
		{
			name: "fedex",
			args: args{
				instant.Data{
					Type: "fedex",
				},
			},
			want: `<img width="12" height="12" alt="fedex" src="/static/favicons/fedex.ico"/> <a href="https://www.fedex.com">FedEx</a>`,
		},
		{
			name: "stackoverflow",
			args: args{
				instant.Data{
					Type: "stackoverflow",
					Solution: instant.StackOverflowAnswer{
						Answer: instant.SOAnswer{
							User: "bob",
						},
					},
				},
			},
			want: `bob via <img width="12" height="12" alt="stackoverflow" src="/static/favicons/stackoverflow.ico"/> <a href="https://stackoverflow.com/">Stack Overflow</a>`,
		},
		{
			name: "stock quote",
			args: args{
				instant.Data{
					Type: "stock quote",
					Solution: &stock.Quote{
						Provider: stock.IEXProvider,
					},
				},
			},
			want: `<img width="12" height="12" alt="iex" src="/static/favicons/iex.ico"/> Data provided for free by <a href="https://iextrading.com/developer">IEX</a>.`,
		},
		{
			name: "ups",
			args: args{
				instant.Data{
					Type: "ups",
				},
			},
			want: `<img width="12" height="12" alt="ups" src="/static/favicons/ups.ico"/> <a href="https://www.ups.com">UPS</a>`,
		},
		{
			name: "usps",
			args: args{
				instant.Data{
					Type: "usps",
				},
			},
			want: `<img width="12" height="12" alt="usps" src="/static/favicons/usps.ico"/> <a href="https://www.usps.com">USPS</a>`,
		},
		{
			name: "weather",
			args: args{
				instant.Data{
					Type: "weather",
					Solution: &weather.Weather{
						Provider: weather.OpenWeatherMapProvider,
					},
				},
			},
			want: `<img width="12" height="12" alt="openweathermap" src="/static/favicons/openweathermap.ico"/> <a href="http://openweathermap.org">OpenWeatherMap</a>`,
		},
		{
			name: "wikidata",
			args: args{
				instant.Data{
					Type: "wikidata",
				},
			},
			want: `<img width="12" height="12" alt="wikipedia" src="/static/favicons/wikipedia.ico"/> <a href="https://www.wikipedia.org/">Wikipedia</a>`,
		},
		{
			name: "wikiquote",
			args: args{
				instant.Data{
					Type: "wikiquote",
				},
			},
			want: `<img width="12" height="12" alt="wikiquote" src="/static/favicons/wikiquote.ico"/> <a href="https://www.wikiquote.org/">Wikiquote</a>`,
		},
		{
			name: "wiktionary",
			args: args{
				instant.Data{
					Type: "wiktionary",
				},
			},
			want: `<img width="12" height="12" alt="wiktionary" src="/static/favicons/wiktionary.ico"/> <a href="https://www.wiktionary.org/">Wiktionary</a>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := source(tt.args.src)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWeatherCode(t *testing.T) {
	for _, tt := range []struct {
		arg  weather.Description
		want string
	}{
		{
			weather.Clear, "icon-sun",
		},
		{
			weather.LightClouds, "icon-cloud-sun",
		},
		{
			weather.ScatteredClouds, "icon-cloud",
		},
		{
			weather.OvercastClouds, "icon-cloud-inv",
		},
		{
			weather.Extreme, "icon-cloud-flash-inv",
		},
		{
			weather.Rain, "icon-rain",
		},
		{
			weather.Snow, "icon-snow",
		},
		{
			weather.ThunderStorm, "icon-cloud-flash",
		},
		{
			weather.Windy, "icon-windy",
		},
		{
			"", "icon-sun",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := weatherCode(tt.arg)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	for _, tt := range []struct {
		s    string
		len  int
		p    bool
		want string
	}{
		{
			s:    "This sentence should be truncated here and not go on and on and on and more on.",
			len:  39,
			p:    true,
			want: "This sentence should be truncated here ...",
		},
		{
			s:    "This sentence should be truncated here and not go on and on and on and more on.",
			len:  30,
			p:    false,
			want: "This sentence should be trunca...",
		},
		{
			s:    "This no truncate",
			len:  25,
			p:    true,
			want: "This no truncate",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := truncate(tt.s, tt.len, tt.p)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestWikiAmount(t *testing.T) {
	type args struct {
		quantity wikipedia.Quantity
		l        language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "meters",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11573"}, Amount: "2.16"},
				language.German,
			},
			want: "2.16 m",
		},
		{
			name: "cm",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q174728"}, Amount: "3"},
				language.French,
			},
			want: "3 cm",
		},
		{
			name: "inches",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q218593"}, Amount: "131"},
				language.French,
			},
			want: "333 cm",
		},
		{
			name: "kg",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"},
				language.Italian,
			},
			want: "147 kg",
		},
		{
			name: "meters (US)",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11573"}, Amount: "2.16"},
				language.English,
			},
			want: `7'1"`,
		},
		{
			name: "cm  (US)",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q174728"}, Amount: "3"},
				language.English,
			},
			want: `1.181103"`,
		},
		{
			name: "kg  (US)",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"},
				language.English,
			},
			want: "324 lbs",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := tt.args.l.Region()

			got := wikiAmount(tt.args.quantity, r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiCanonical(t *testing.T) {
	type args struct {
		title string
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{"jimi hendrix was here"},
			want: "jimi_hendrix_was_here",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiCanonical(tt.args.title)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiData(t *testing.T) {
	type args struct {
		sol instant.Data
		l   language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "empty",
			args: args{
				instant.Data{
					Solution: []wikipedia.Quantity{},
				},
				language.English,
			},
			want: "",
		},
		{
			name: "kg",
			args: args{
				instant.Data{
					Solution: []wikipedia.Quantity{{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"}},
				},
				language.Italian,
			},
			want: "147 kg",
		},
		{
			name: "age (alive)",
			args: args{
				instant.Data{
					Solution: instant.Age{
						Birthday: instant.Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1972-12-31T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
				language.English,
			},
			want: `<em>Age:</em> 45 Years<br><span style="color:#666;">December 31, 1972</span>`,
		},
		{
			name: "age (at time of death)",
			args: args{
				instant.Data{
					Solution: instant.Age{
						Birthday: instant.Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1956-04-30T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
						Death: instant.Death{
							Death: wikipedia.DateTime{
								Value:    "1984-03-13T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
				language.English,
			},
			want: `<em>Age at Death:</em> 27 Years<br><span style="color:#666;">April 30, 1956 - March 13, 1984</span>`,
		},
		{
			name: "birthday",
			args: args{
				instant.Data{
					Solution: instant.Birthday{
						Birthday: wikipedia.DateTime{
							Value:    "2001-05-14T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
				language.English,
			},
			want: `May 14, 2001`,
		},
		{
			name: "death",
			args: args{
				instant.Data{
					Solution: instant.Death{
						Death: wikipedia.DateTime{
							Value:    "2015-05-14T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
				language.English,
			},
			want: `May 14, 2015`,
		},
		{
			name: "unknown",
			args: args{
				instant.Data{Solution: 1}, language.English},
			want: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := tt.args.l.Region()
			now = func() time.Time {
				return time.Date(2018, 02, 06, 20, 34, 58, 651387237, time.UTC)
			}

			got := wikiData(tt.args.sol, r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWikiDateTime(t *testing.T) {
	type args struct {
		dt wikipedia.DateTime
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "birthday",
			args: args{
				wikipedia.DateTime{
					Value:    "1972-12-31T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: "December 31, 1972",
		},
		{
			name: "year",
			args: args{
				wikipedia.DateTime{
					Value:    "1987",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: "1987",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiDateTime(tt.args.dt)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiJoin(t *testing.T) {
	type args struct {
		items     []wikipedia.Wikidata
		preferred []language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{
				[]wikipedia.Wikidata{
					{ID: "1", Labels: map[string]wikipedia.Text{
						"fr": {Text: "rock in french", Language: "fr"},
						"en": {Text: "rock", Language: "en"},
					}},
					{ID: "1", Labels: map[string]wikipedia.Text{
						"en": {Text: "rap", Language: "en"},
						"de": {Text: "rap in german", Language: "de"},
					}},
					{ID: "1", Labels: map[string]wikipedia.Text{
						"en": {Text: "country", Language: "en"},
					}},
				},
				[]language.Tag{
					language.English, language.French, language.German,
				},
			},
			want: "rock, rap, country",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiJoin(tt.args.items, tt.args.preferred)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiLabel(t *testing.T) {
	type args struct {
		labels    map[string]wikipedia.Text
		preferred []language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{
				map[string]wikipedia.Text{
					"en":    {Text: "english language", Language: "en"},
					"de":    {Text: "german language", Language: "de"},
					"sr-el": {Text: "this doesn't parse language", Language: "sr-el"},
				},
				[]language.Tag{
					language.English, language.French, language.German,
				},
			},
			want: "english language",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiLabel(tt.args.labels, tt.args.preferred)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikipediaItem(t *testing.T) {
	want := &wikipedia.Item{}

	d := instant.Data{
		Solution: &wikipedia.Item{},
	}

	got := wikipediaItem(d)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestWikiYears(t *testing.T) {
	type args struct {
		start interface{}
		end   interface{}
	}

	for _, tt := range []struct {
		name string
		args
		want int
	}{
		{
			name: "zero",
			args: args{
				time.Time{},
				time.Time{},
			},
			want: 0,
		},
		{
			name: "basic",
			args: args{
				time.Date(1975, 11, 17, 20, 34, 58, 651387237, time.UTC),
				time.Date(2017, 11, 18, 20, 34, 58, 651387237, time.UTC),
			},
			want: 42,
		},
		{
			name: "almost 42",
			args: args{
				time.Date(1975, 11, 17, 20, 34, 58, 651387237, time.UTC),
				time.Date(2017, 11, 16, 20, 34, 58, 651387237, time.UTC),
			},
			want: 41,
		},
		{
			name: "wikiDateTime",
			args: args{
				wikipedia.DateTime{
					Value:    "1854-12-31T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
				wikipedia.DateTime{
					Value:    "1912-04-30T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: 57,
		},
		{
			name: "wikiDateTime year",
			args: args{
				wikipedia.DateTime{
					Value:    "1794",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
				wikipedia.DateTime{
					Value:    "1954-02-14T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: 160,
		},
		{
			name: "wrong type",
			args: args{5, 12},
			want: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiYears(tt.args.start, tt.args.end)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
