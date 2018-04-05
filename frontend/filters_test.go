package frontend

import (
	"html/template"
	"reflect"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant/parcel"
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
		number int64
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
			number: 999000,
			want:   "999,000",
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

func TestInstantFormatter(t *testing.T) {
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
			name: "string",
			args: args{
				instant.Data{
					Solution: "basic string",
				},
				language.English,
			},
			want: `basic string`,
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
							Value:    "1938-07-31T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
				language.English,
			},
			want: `July 31, 1938`,
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
			name: "wikiquote",
			args: args{
				instant.Data{
					Solution: []string{"fantastic quote", "such good quote"},
				},
				language.English,
			},
			want: `<p><span style="font-size:14px;font-style:italic;">fantastic quote</span></p><p><span style="font-size:14px;font-style:italic;">such good quote</span></p>`,
		},
		{
			name: "wiktionary",
			args: args{
				instant.Data{
					Solution: wikipedia.Wiktionary{
						Title:    "guitar",
						Language: "en",
						Definitions: []*wikipedia.Definition{
							{
								Part:    "noun",
								Meaning: "an instrument",
								Synonyms: []wikipedia.Synonym{
									{
										Word:     "axe",
										Language: "en",
									},
								},
							},
						},
					},
				},
				language.English,
			},
			want: `<p><span style="font-size:18px;"><em><a href="https://en.wiktionary.org/wiki/guitar" style="color:#333;">guitar</a></em></span></p><span style="font-size:14px;font-style:italic;">noun</span><br><span style="display:inline-block;margin-left:15px;">an instrument</span><br><span style="display:inline-block;margin-left:15px;font-style:italic;color:#666;">synonyms:&nbsp;</span><a href="https://en.wiktionary.org/wiki/axe" >axe</a><br><br>`,
		},
		{
			name: "unknown",
			args: args{
				instant.Data{Solution: 1}, language.English},
			want: "",
		},
		{
			name: "stock quote",
			args: args{
				instant.Data{
					Type: "stock quote",
					Solution: &stock.Quote{
						Ticker:   "TCKR",
						Name:     "Some Company",
						Exchange: stock.NYSE,
						Last: stock.Last{
							Price:         12.43,
							Time:          time.Date(2018, 3, 3, 9, 45, 42, 0, time.UTC),
							Change:        -.423,
							ChangePercent: -.0103,
						},
						History: []stock.EOD{
							{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
						},
					},
				},
				language.English,
			},
			want: `<div class="pure-u-1"><div class="pure-u-1" style="font-size:20px;">Some Company</div><div class="pure-u-1" style="font-size:14px;">NYSE: TCKR <span id="quote_time" style="font-size:12px;">March 3, 2018 9:45 AM UTC</span></div></div><div class="pure-u-1" style="font-size:40px;">12.43 <span style="font-size:22px;"><span class="quote-arrow quote-arrow-down"></span><span style="color:#C80000;"> -0.42 (-1.03%)</span></span></div><div id="stock_chart" class="pure-u-1"></div><div class="pure-u-1"><div id="time_period_buttons" class="pure-button-group" role="group" aria-label="time select" style="margin-left:47px;"><button id="day" class="pure-button" disabled>Day</button>&nbsp;&nbsp;<button id="week" class="pure-button">Week</button>&nbsp;&nbsp;<button id="month" class="pure-button">Month</button>&nbsp;&nbsp;<button id="ytd" class="pure-button">YTD</button>&nbsp;&nbsp;<button id="1yr" class="pure-button">1 Year</button>&nbsp;&nbsp;<button id="5yr" class="pure-button">5 Year</button></div></div><script>var data = [{"date":"2013-03-26T00:00:00Z","open":60.5276,"close":59.9679,"high":60.5797,"low":59.8891,"volume":73428208}];</script>`,
		},
		{
			name: "weather",
			args: args{
				instant.Data{
					Type: "weather",
					Solution: &weather.Weather{
						City: "Centerville",
						Today: weather.Today{
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
							Low:         55.4,
							High:        62.6,
						},
						Provider: weather.OpenWeatherMapProvider,
					},
				},
				language.English,
			},
			want: `<div class="pure-u-1"><div class="pure-u-1" style="margin-bottom:15px;font-size:18px;text-shadow:rgba(0,0,0,.3);">Centerville</div><div class="pure-u-1" style="vertical-align:top;"><i class="icon-cloud icon-large" aria-hidden="true" style="text-shadow:1px 1px 1px #ccc;vertical-align:top;margin-right:10px;"></i><span style="font-size:48px;font-weight:200;text-shadow:rgba(0,0,0,.3);cursor:default;">59</span><span style="width:25px;display:inline-block;vertical-align:top;margin-top:5px;"><i class="icon-fahrenheit" aria-hidden="true"></i><hr style="display:none;"><i class="icon-celsius" aria-hidden="true" style="display:none;"></i></span><span style="display:inline-block;vertical-align:top;margin-top:14px;margin-left:25px;"><em>H</em> 62.6&deg;<hr style="opacity:0;"><em>L</em> 55.4&deg;</span><span style="display:inline-block;vertical-align:top;margin-left:25px;"><hr style="opacity:0;"><em>Wind:</em> 4.7 MPH<hr style="opacity:0;"><em>Humidity:</em> 33%<hr style="opacity:0;"><em>Clouds:</em> 40%</span></div></div>`,
		},
		{
			name: "tracking package",
			args: args{
				instant.Data{
					Type: "ups",
					Solution: parcel.Response{
						TrackingNumber: "90210",
						Updates: []parcel.Update{
							{
								DateTime: time.Date(2017, 2, 2, 15, 34, 59, 1, time.UTC),
								Location: parcel.Location{
									City:    "Armadillo",
									State:   "TX",
									Country: "US",
								},
								Status: "Departure Scan",
							},
						},
						Expected: parcel.Expected{
							Delivery: "Scheduled Delivery",
							Date:     time.Date(2017, 2, 2, 15, 34, 59, 1, time.UTC),
						},
						URL: "https://www.ups.com/some/random/url?and=query",
					},
				},
				language.English,
			},
			want: `<img width="18" height="18" alt="ups" src="/static/favicons/ups.ico" style="vertical-align:middle"/> <a href="https://www.ups.com/some/random/url?and=query"><em>90210</em></a><br><p><span style="font-weight:bold;font-size:20px;">Scheduled Delivery: Thursday, February 2, 2017</span></p><div class="pure-u-1" style="margin-bottom:5px;"><div class="pure-u-7-24" style="font-weight:bold;">DATE</div><div class="pure-u-9-24" style="font-weight:bold;">LOCATION</div><div class="pure-u-8-24" style="font-weight:bold;">STATUS</div></div><div class="pure-u-1" style="color:#444;font-size:14px;margin-bottom:10px;"><div class="pure-u-7-24">Thu, 02 Feb 3:34PM</div><div class="pure-u-9-24">Armadillo, TX, US</div><div class="pure-u-8-24">Departure Scan</div></div>`,
		},
		{
			name: "delivered package",
			args: args{
				instant.Data{
					Type: "ups",
					Solution: parcel.Response{
						TrackingNumber: "SomeTrackingNumber",
						Updates: []parcel.Update{
							{
								DateTime: time.Date(2017, 2, 3, 15, 34, 45, 1, time.UTC),
								Location: parcel.Location{
									City:    "Final Destination Yo!",
									State:   "UT",
									Country: "US",
								},
								Status: "Delivered",
							},
							{
								DateTime: time.Date(2017, 2, 2, 15, 34, 59, 1, time.UTC),
								Location: parcel.Location{
									City:    "Armadillo",
									State:   "TX",
									Country: "US",
								},
								Status: "Departure Scan",
							},
						},
						Expected: parcel.Expected{},
						URL:      "https://www.ups.com/some/random/url?and=query",
					},
				},
				language.English,
			},
			want: `<img width="18" height="18" alt="ups" src="/static/favicons/ups.ico" style="vertical-align:middle"/> <a href="https://www.ups.com/some/random/url?and=query"><em>SomeTrackingNumber</em></a><br><p><span style="font-weight:bold;font-size:20px;">Delivered: Friday, February 3, 2017 3:34PM</span></p><div class="pure-u-1" style="margin-bottom:5px;"><div class="pure-u-7-24" style="font-weight:bold;">DATE</div><div class="pure-u-9-24" style="font-weight:bold;">LOCATION</div><div class="pure-u-8-24" style="font-weight:bold;">STATUS</div></div><div class="pure-u-1" style="color:#444;font-size:14px;margin-bottom:10px;"><div class="pure-u-7-24">Fri, 03 Feb 3:34PM</div><div class="pure-u-9-24">Final Destination Yo!, UT, US</div><div class="pure-u-8-24">Delivered</div></div><div class="pure-u-1" style="color:#444;font-size:14px;margin-bottom:10px;"><div class="pure-u-7-24">Thu, 02 Feb 3:34PM</div><div class="pure-u-9-24">Armadillo, TX, US</div><div class="pure-u-8-24">Departure Scan</div></div>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := tt.args.l.Region()
			now = func() time.Time {
				return time.Date(2018, 02, 06, 20, 34, 58, 651387237, time.UTC)
			}

			got := instantFormatter(tt.args.sol, r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
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
