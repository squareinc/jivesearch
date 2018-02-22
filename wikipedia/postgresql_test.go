package wikipedia

import (
	"reflect"
	"testing"

	"golang.org/x/text/language"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var shaqClaimsJSON = []byte(`{
	"sex": [{"id": "Q6581097", "labels": {"en": {"value": "male", "language": "en"}}}]
}`)

var shaqClaimsPostgres = &Claims{
	Sex: []Wikidata{
		{
			ID: "Q6581097",
			Labels: map[string]Text{
				"en": {Text: "male", Language: "en"},
			},
			Claims: &Claims{},
		},
	},
}

func TestPostgreSQL_Fetch(t *testing.T) {
	type args struct {
		query string
		lang  language.Tag
	}
	tests := []struct {
		name string
		args args
		want *Item
	}{
		{
			"shaq",
			args{"Shaquille O'Neal", language.MustParse("en")},
			&Item{
				Wikipedia: Wikipedia{
					Language: "en",
					Title:    "Shaquille O'Neal",
					Text:     "Shaquille O'Neal is a basketball player",
				},
				Wikidata: &Wikidata{
					ID:           "Q169452",
					Descriptions: shaqDescriptions,
					Aliases:      shaqAliases,
					Labels:       shaqLabels,
					Claims:       shaqClaimsPostgres,
				},
				Wikiquote: Wikiquote{
					Quotes: []string{},
					//Quotes: shaqQuotes, // not really sure how to test for pq.Array
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			rows := sqlmock.NewRows(
				[]string{"id", "title", "text", "quotes", "labels", "aliases", "descriptions", "claims"},
			)
			rows = rows.AddRow(
				"Q169452", tt.args.query, "Shaquille O'Neal is a basketball player", "{}",
				[]byte(shaqRawLabels), []byte(shaqRawAliases), []byte(shaqRawDescriptions), shaqClaimsJSON,
			)

			mock.ExpectQuery("SELECT").WithArgs(tt.args.query).WillReturnRows(rows)

			p := &PostgreSQL{
				DB: db,
			}

			got, err := p.Fetch(tt.args.query, tt.args.lang)
			if err != nil {
				t.Fatal(err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQL_Dump(t *testing.T) {
	type args struct {
		lang language.Tag
		ft   FileType
	}
	tests := []struct {
		name string
		row  interface{}
		args args
	}{
		{
			"enwiki",
			shaqWikipedia,
			args{
				ft:   WikipediaFT,
				lang: language.MustParse("en"),
			},
		},
		{
			"wikidata",
			shaqWikidata,
			args{
				ft: WikidataFT,
			},
		},
		{
			"enwikiquote",
			shaqWikiquote,
			args{
				ft:   WikiquoteFT,
				lang: language.MustParse("en"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			// create table
			mock.ExpectExec("DROP TABLE IF EXISTS").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(1, 1))

			// insert data
			mock.ExpectBegin()
			mock.ExpectPrepare("COPY")
			mock.ExpectCommit()

			// create indices
			mock.ExpectBegin()

			switch tt.args.ft {
			case WikipediaFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikidataFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikiquoteFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			}

			mock.ExpectCommit()

			// rename table
			mock.ExpectBegin()
			mock.ExpectExec("DROP TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("ALTER TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			switch tt.args.ft {
			case WikipediaFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikidataFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikiquoteFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			}
			mock.ExpectCommit()

			p := &PostgreSQL{
				DB: db,
			}

			rows := make(chan interface{})

			go func() {
				rows <- tt.row
			}()

			if err := p.Dump(tt.args.ft, tt.args.lang, rows); err != nil {
				t.Fatal(err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &PostgreSQL{
		DB: db,
	}

	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))

	err = p.Setup()
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
