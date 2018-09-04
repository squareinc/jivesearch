package congress

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestFetchSenators(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		key   string
		state string
	}

	for _, tt := range []struct {
		name string
		args
		u    string
		resp string
		want *Response
	}{
		{
			name: "basic",
			u:    `https://api.propublica.org/congress/v1/members/senate/UT/current.json`,
			resp: `{"status":"OK","copyright":"Copyright (c) 2018 Pro Publica Inc. All Rights Reserved.","results":[{"id":"H000338","name":"Orrin G. Hatch","first_name":"Orrin","middle_name":"G.","last_name":"Hatch","suffix":null,"role":"Senator, 1st Class","gender":"M","party":"R","times_topics_url":"http:\/\/topics.nytimes.com\/top\/reference\/timestopics\/people\/h\/orrin_g_hatch\/index.html","twitter_id":"SenOrrinHatch","facebook_account":"senatororrinhatch","youtube_id":"SenatorOrrinHatch","seniority":"41","next_election":"2018","api_uri":"https:\/\/api.propublica.org\/congress\/v1\/members\/H000338.json"},{"id":"L000577","name":"Mike Lee","first_name":"Mike","middle_name":null,"last_name":"Lee","suffix":null,"role":"Senator, 3rd Class","gender":"M","party":"R","times_topics_url":"","twitter_id":"SenMikeLee","facebook_account":"senatormikelee","youtube_id":"senatormikelee","seniority":"7","next_election":"2022","api_uri":"https:\/\/api.propublica.org\/congress\/v1\/members\/L000577.json"}]}`,
			args: args{
				key:   "some_key",
				state: "utah",
			},
			want: &Response{
				Location: &Location{
					Short: "UT",
					State: "Utah",
				},
				Role: Senators,
				Members: []Member{
					{
						Name:         "Orrin G. Hatch",
						Gender:       "M",
						Party:        "R",
						Twitter:      "SenOrrinHatch",
						Facebook:     "senatororrinhatch",
						NextElection: 2018,
					},
					{
						Name:         "Mike Lee",
						Gender:       "M",
						Party:        "R",
						Twitter:      "SenMikeLee",
						Facebook:     "senatormikelee",
						NextElection: 2022,
					},
				},
				Provider: ProPublicaProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder) // no responder found????

			p := &ProPublica{
				Key:        tt.args.key,
				HTTPClient: &http.Client{},
			}

			loc := ValidateState(tt.args.state)

			got, err := p.FetchSenators(loc)
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
