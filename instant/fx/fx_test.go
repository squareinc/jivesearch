package fx

import (
	"reflect"
	"testing"
)

func TestSetBase(t *testing.T) {
	for _, tt := range []struct {
		name string
		args *Rate
		want *Rate
	}{
		{
			name: "basic",
			args: &Rate{Currency: EUR},
			want: &Rate{Base: USD, Currency: EUR},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.setBase()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
