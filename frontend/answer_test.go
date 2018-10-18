package frontend

import (
	"reflect"
	"testing"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
)

func TestDetectType(t *testing.T) {
	for _, c := range []struct {
		name instant.Type
		want interface{}
	}{
		{instant.BirthStoneType, nil},
		{instant.BreachType, &breach.Response{}},
		{instant.CountryCodeType, &instant.CountryCodeResponse{}},
		{instant.CurrencyType, &instant.CurrencyResponse{}},
		{instant.DiscographyType, &[]discography.Album{}},
		{instant.FedExType, &parcel.Response{}},
		{instant.GDPType, &instant.GDPResponse{}},
		{instant.HashType, &instant.HashResponse{}},
		{instant.PopulationType, &instant.PopulationResponse{}},
		{instant.StackOverflowType, &instant.StackOverflowAnswer{}},
		{instant.StockQuoteType, &stock.Quote{}},
		{instant.URLShortenerType, &shortener.Response{}},
		{instant.WeatherType, &weather.Weather{}},
		{instant.WikipediaType, []*wikipedia.Item{}},
		{
			"wikidata age", &instant.Age{
				Birthday: &instant.Birthday{},
				Death:    &instant.Death{},
			},
		},
		{instant.WikidataBirthdayType, &instant.Birthday{}},
		{instant.WikidataDeathType, &instant.Death{}},

		{instant.WikidataHeightType, &[]wikipedia.Quantity{}},
		{instant.WikiquoteType, &[]string{}},
		{instant.WiktionaryType, &wikipedia.Wiktionary{}},
	} {
		t.Run(string(c.name), func(t *testing.T) {
			got := detectType(c.name)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}
