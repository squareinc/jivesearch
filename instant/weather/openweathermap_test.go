package weather

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestOpenWeatherMapFetchByZip(t *testing.T) {
	type args struct {
		zip int
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		u    string
		resp string
		want *Weather
	}{
		{
			name: "Bountiful, Utah",
			args: args{84014},
			u:    `https://api.openweathermap.org/data/2.5/weather?APPID=myappid&zip=84014,us&units=imperial`,
			resp: `{"coord":{"lon":-111.88,"lat":40.89},"weather":[{"id":802,"main":"Clouds","description":"scattered clouds","icon":"03d"}],"base":"stations","main":{"temp":58.53,"pressure":1014,"humidity":33,"temp_min":55.4,"temp_max":62.6},"visibility":16093,"wind":{"speed":4.7},"clouds":{"all":40},"dt":1522609080,"sys":{"type":1,"id":2802,"message":0.004,"country":"US","sunrise":1522588167,"sunset":1522634001},"id":5771826,"name":"Bountiful","cod":200}`,
			want: &Weather{
				City: "Bountiful",
				Today: Today{
					Code:        ScatteredClouds,
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
				Provider: OpenWeatherMapProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			o := &OpenWeatherMap{
				HTTPClient: &http.Client{},
				Key:        "myappid",
			}
			got, err := o.FetchByZip(tt.args.zip)
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

func TestConvertCode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		c    int
		want weatherCode
	}{
		{200, ThunderStorm},
		{300, Rain},
		{500, Rain},
		{600, Snow},
		{700, Extreme},
		{800, Clear},
		{801, LightClouds},
		{804, OvercastClouds},
		{900, Extreme},
		{951, Windy},
	} {
		t.Run(string(tt.c), func(t *testing.T) {
			r := &openWeatherMapResponse{
				Weather: &Weather{},
			}

			if err := r.convertCode(tt.c); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(r.Weather.Today.Code, tt.want) {
				t.Errorf("got %+v, want %+v", r.Weather.Today.Code, tt.want)
			}
		})
	}

	httpmock.Reset()
}
