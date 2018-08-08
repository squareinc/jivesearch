package fx

import (
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	for _, tt := range []struct {
		name string
		want *Response
	}{
		{
			name: "basic",
			want: &Response{
				Base:       USD,
				Currencies: Currencies,
				History:    make(map[Currency][]*Rate),
				Provider:   ECBProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := New()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSort(t *testing.T) {
	for _, tt := range []struct {
		name string
		args *Response
		want *Response
	}{
		{
			name: "basic",
			args: &Response{
				Base: USD,
				History: map[Currency][]*Rate{
					JPY: {
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.1,
						},
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.12,
						},
					},
					GBP: {
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.5,
						},
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.6,
						},
					},
				},
				Provider: ECBProvider,
			},
			want: &Response{
				Base: USD,
				History: map[Currency][]*Rate{
					JPY: {
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.12,
						},
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.1,
						},
					},
					GBP: {
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.5,
						},
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.6,
						},
					},
				},
				Provider: ECBProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Sort()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
