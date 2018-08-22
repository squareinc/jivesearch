package population

// can't seem to figure out the xml mocking...keep getting "no responder found"
/*
func TestFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		country string
		from    time.Time
		to      time.Time
	}

	for _, tt := range []struct {
		name string
		args
		u    map[string]string
		want *Response
	}{
		{
			name: "basic",
			u: map[string]string{
				`http://api.worldbank.org/v2/countries/it/indicators/SP.POP.TOTL?date=1930%3A2018`: `<?xml version="1.0" encoding="utf-8"?><wb:data page="1" pages="2" per_page="50" total="58" lastupdated="2018-07-25" xmlns:wb="http://www.worldbank.org"><wb:data><wb:indicator id="SP.POP.TOTL">Population, total</wb:indicator><wb:country id="IT">Italy</wb:country><wb:countryiso3code>ITA</wb:countryiso3code><wb:date>2017</wb:date><wb:value>18</wb:value> <wb:unit/> <wb:obs_status/> <wb:decimal>0</wb:decimal></wb:data><wb:data><wb:indicator id="SP.POP.TOTL">Population, total</wb:indicator><wb:country id="IT">Italy</wb:country><wb:countryiso3code>ITA</wb:countryiso3code><wb:date>2003</wb:date><wb:value>2</wb:value> <wb:unit/> <wb:obs_status/> <wb:decimal>0</wb:decimal></wb:data><wb:data><wb:indicator id="SP.POP.TOTL">Population, total</wb:indicator><wb:country id="IT">Italy</wb:country><wb:countryiso3code>ITA</wb:countryiso3code><wb:date>1994</wb:date><wb:value>4</wb:value> <wb:unit/> <wb:obs_status/> <wb:decimal>0</wb:decimal></wb:data></wb:data>`,
			},
			args: args{
				country: "IT",
				from:    time.Date(1930, 12, 31, 0, 0, 0, 0, time.UTC),
				to:      time.Date(2018, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			want: &Response{
				History: []Instant{
					{time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC), 18},
					{time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC), 2},
					{time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC), 4},
				},
				Provider: TheWorldBankProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.u {
				x := &WorldBankResponse{}
				if err := xml.Unmarshal([]byte(v), x); err != nil {
					t.Fatal(err)
				}

				responder, err := httpmock.NewXmlResponder(200, x)
				if err != nil {
					t.Fatal(err)
				}

				httpmock.RegisterResponder("GET", k, responder) // no responder found????
			}

			w := &WorldBank{
				HTTPClient: &http.Client{},
			}

			got, err := w.Fetch(tt.args.country, tt.args.from, tt.args.to)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}
*/
