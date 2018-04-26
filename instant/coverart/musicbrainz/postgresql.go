package musicbrainz

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/lib/pq"

	"github.com/jivesearch/jivesearch/instant/coverart"
)

// PostgreSQL holds configuration for a MusicBrainz database
type PostgreSQL struct {
	*sql.DB
}

// Fetch fetches links to MusicBrainz cover art from Postgresql
// Note: The Cover Art Archive api has no rate limits but is a bit slow.
func (p *PostgreSQL) Fetch(ids []string) (map[string]coverart.Image, error) {
	imgs := map[string]coverart.Image{}
	var err error

	// https://github.com/metabrainz/coverart_redirect/blob/master/coverart_redirect/request.py
	// get the release GID that has a cover from the release group GID
	sql := `
		SELECT DISTINCT ON (release.release_group) release.gid AS mbid, release_group.gid
		FROM musicbrainz.index_listing
		JOIN musicbrainz.release
			ON musicbrainz.release.id = musicbrainz.index_listing.release
		JOIN musicbrainz.release_group
			ON release_group.id = release.release_group
		LEFT JOIN (
			SELECT release, date_year, date_month, date_day
			FROM musicbrainz.release_country
			UNION ALL
			SELECT release, date_year, date_month, date_day
			FROM musicbrainz.release_unknown_country
		) release_event ON (release_event.release = release.id)
		FULL OUTER JOIN musicbrainz.release_group_cover_art
		ON release_group_cover_art.release = musicbrainz.release.id
		WHERE release_group.gid = ANY ($1)
		AND is_front = true
		ORDER BY release.release_group, release_group_cover_art.release,
			release_event.date_year, release_event.date_month,
			release_event.date_day
	`

	rows, err := p.DB.Query(sql, pq.Array(ids))
	if err != nil {
		return imgs, err
	}

	m := map[string]string{}
	var mbids = []string{}

	defer rows.Close()
	for rows.Next() {
		var mbid string
		var gid string
		if err := rows.Scan(&mbid, &gid); err != nil {
			return imgs, err
		}

		m[mbid] = gid
		mbids = append(mbids, mbid)
	}

	if err := rows.Err(); err != nil {
		return imgs, err
	}

	sql = `
		SELECT musicbrainz.release.gid, index_listing.id, image_type.suffix
		FROM musicbrainz.index_listing
		JOIN musicbrainz.release
		ON musicbrainz.index_listing.release = musicbrainz.release.id
		JOIN musicbrainz.image_type
		ON musicbrainz.index_listing.mime_type = musicbrainz.image_type.mime_type
		WHERE musicbrainz.release.gid = ANY ($1)
		AND is_front = true
		ORDER BY ordering
	`

	rows, err = p.DB.Query(sql, pq.Array(mbids))
	if err != nil {
		return imgs, err
	}

	defer rows.Close()
	for rows.Next() {
		var mbid string
		var id string
		var ext string

		if err := rows.Scan(&mbid, &id, &ext); err != nil {
			return imgs, err
		}

		u, _ := url.Parse(fmt.Sprintf("http://coverartarchive.org/release/%v/%v-250.%v", mbid, id, ext))
		img := coverart.Image{
			ID:          m[mbid],
			URL:         u,
			Description: coverart.Front,
			Width:       250,
			Height:      250,
		}

		imgs[m[mbid]] = img
	}

	if err := rows.Err(); err != nil {
		return imgs, err
	}

	return imgs, err
}
