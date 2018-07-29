package fx

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestECBFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
		want *Response
	}{
		{
			name: "basic",
			u:    `http://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml`,
			resp: `<?xml version="1.0" encoding="UTF-8"?><gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref"><gesmes:subject>Reference rates</gesmes:subject><gesmes:Sender><gesmes:name>European Central Bank</gesmes:name></gesmes:Sender><Cube><Cube time='2018-07-27'> <Cube currency='USD' rate='1.1625'/> <Cube currency='JPY' rate='129.25'/> <Cube currency='BGN' rate='1.9558'/> </Cube></Cube></gesmes:Envelope>`,
			want: &Response{
				Rates: []*Rate{
					{USD, BGN, .5944},
					{USD, EUR, 1.1625},
					{USD, JPY, 0.009},
					{USD, USD, 1.0},
				},
				DateTime: time.Date(2018, 07, 27, 0, 0, 0, 0, time.UTC),
				Provider: ECBProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			ecb := &ECB{}
			got, err := ecb.Fetch()
			if err != nil {
				t.Fatal(err)
			}

			sort.Slice(got.Rates, func(i, j int) bool {
				return got.Rates[i].Currency.Short < got.Rates[j].Currency.Short
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}
