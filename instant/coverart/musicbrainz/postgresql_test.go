package musicbrainz

import (
	"database/sql/driver"
	"net/url"
	"reflect"
	"testing"

	"github.com/lib/pq"

	"github.com/jivesearch/jivesearch/instant/coverart"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestFetch(t *testing.T) {
	type args struct {
		ids  []string
		row1 []driver.Value
		row2 []driver.Value
	}

	u, _ := url.Parse("http://coverartarchive.org/release/1/2-250..jpg")

	tests := []struct {
		name string
		args args
		want map[string]coverart.Image
	}{
		{
			"basic",
			args{
				[]string{"1"},
				[]driver.Value{"1", "2"},
				[]driver.Value{"1", "2", ".jpg"},
			},
			map[string]coverart.Image{
				"2": {
					ID:          "2",
					URL:         u,
					Description: coverart.Front,
					Height:      250,
					Width:       250,
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
				[]string{"mbid", "gid"},
			)

			rows = rows.AddRow(
				tt.args.row1...,
			)

			mock.ExpectQuery("SELECT").WithArgs(pq.Array(tt.args.ids)).WillReturnRows(rows)

			rows = sqlmock.NewRows(
				[]string{"gid", "id", "suffix"},
			)
			rows = rows.AddRow(
				tt.args.row2...,
			)

			mock.ExpectQuery("SELECT").WithArgs(pq.Array(tt.args.ids)).WillReturnRows(rows)

			p := &PostgreSQL{
				DB: db,
			}

			got, err := p.Fetch(tt.args.ids)
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
