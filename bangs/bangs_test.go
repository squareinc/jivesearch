package bangs

import (
	"testing"

	"golang.org/x/text/language"
)

// TestDefault tests that each !bang has a default location
func TestDefault(t *testing.T) {
	b := New()
	for _, bng := range b.Bangs {
		if _, ok := bng.Regions[def]; !ok {
			t.Fatalf("%q bang needs a default region", bng.Name)
		}
	}
}

func TestDetect(t *testing.T) {
	type data struct {
		loc string
		ok  bool
	}

	for _, c := range []struct {
		q    string
		r    string
		l    language.Tag
		want data
	}{
		{
			q: "!g bob", r: "US", l: language.French,
			want: data{
				loc: "https://encrypted.google.com/search?hl=fr&q=bob",
				ok:  true,
			},
		},
		{
			q: "!g bob french", r: "fr", l: language.English,
			want: data{
				loc: "https://www.google.fr/search?hl=en&q=bob french",
				ok:  true,
			},
		},
		{
			q: "!gfr something french", r: "fr", l: language.English,
			want: data{
				loc: "https://www.google.fr/search?hl=en&q=something french",
				ok:  true,
			},
		},
		{
			q: "!W bob maRLey", r: "US", l: language.French,
			want: data{
				loc: "https://en.wikipedia.org/wiki/Bob_Marley",
				ok:  true,
			},
		},
		{
			q: "nonexistent! some query", r: "US", l: language.French,
			want: data{
				loc: "",
				ok:  false,
			},
		},
		{
			q: "this is not a bang", r: "US", l: language.English,
			want: data{
				loc: "",
				ok:  false,
			},
		},
		{
			q: "this is not a bang g", r: "US", l: language.English,
			want: data{
				loc: "",
				ok:  false,
			},
		},
		{
			q: "this is not a bang google", r: "US", l: language.English,
			want: data{
				loc: "",
				ok:  false,
			},
		},
	} {
		t.Run(c.q, func(t *testing.T) {
			b := New()

			r := language.MustParseRegion(c.r)

			var got = data{}
			got.loc, got.ok = b.Detect(c.q, r, c.l)
			if got != c.want {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}
