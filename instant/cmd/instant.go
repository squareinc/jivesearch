// Sample instant demonstrates how to run a simple instant answers server.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/abursavich/nett"
	"github.com/jivesearch/jivesearch/config"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

type cfg struct {
	*instant.Instant
}

func (c *cfg) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sol := c.Instant.Detect(r, language.English)

	if err := json.NewEncoder(w).Encode(sol); err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
}

func favHandler(w http.ResponseWriter, r *http.Request) {}

func setup() (*sql.DB, string, error) {
	v := viper.New()
	v.SetEnvPrefix("jivesearch")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetDefaults(v)

	db, err := sql.Open("postgres",
		fmt.Sprintf(
			"user=%s password=%s host=%s database=%s sslmode=require",
			v.GetString("postgresql.user"),
			v.GetString("postgresql.password"),
			v.GetString("postgresql.host"),
			v.GetString("postgresql.database"),
		),
	)
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(0)

	return db, v.GetString("stackoverflow.key"), err
}

func main() {
	db, key, err := setup()
	if err != nil {
		panic(err)
	}

	defer db.Close()

	c := &cfg{
		&instant.Instant{
			QueryVar: "q",
			StackOverflowFetcher: &stackoverflow.API{
				Key: key,
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						Dial: (&nett.Dialer{
							Resolver: &nett.CacheResolver{TTL: 10 * time.Minute},
							IPFilter: nett.DualStack,
						}).Dial,
						DisableKeepAlives: true,
					},
					Timeout: 5 * time.Second,
				},
			},
			WikipediaFetcher: &wikipedia.PostgreSQL{
				DB: db,
			},
		},
	}

	if err := c.Instant.WikipediaFetcher.Setup(); err != nil {
		panic(err)
	}

	port := 8000
	http.HandleFunc("/", c.handler)
	http.HandleFunc("/favicon.ico", favHandler)
	log.Printf("Listening at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
