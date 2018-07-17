// Package frontend provides the routing and middleware for the web app
package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/bangs"
	"github.com/jivesearch/jivesearch/frontend/cache"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/log"
	"github.com/jivesearch/jivesearch/search"
	img "github.com/jivesearch/jivesearch/search/image"
	"github.com/jivesearch/jivesearch/search/vote"
	"github.com/jivesearch/jivesearch/suggest"
	"github.com/oxtoacart/bpool"
	"golang.org/x/text/language"
)

// Frontend holds settings for branding, cache, search backend, etc.
type Frontend struct {
	Brand
	Document
	*bangs.Bangs
	Cache struct {
		cache.Cacher
		Instant time.Duration
		Search  time.Duration
	}
	Images struct {
		img.Fetcher
		*http.Client
	}
	*instant.Instant
	MapBoxKey string
	Suggest   suggest.Suggester
	Search    search.Fetcher
	Host      string
	Wikipedia
	Vote vote.Voter
	GitHub
}

// Brand allows for customization of the name and tagline
type Brand struct {
	Name      string
	TagLine   string
	Logo      string
	SmallLogo string
}

// Document has the languages we support
type Document struct {
	Languages []language.Tag
	language.Matcher
}

// Wikipedia holds our settings for wikipedia/wikidata
// Note: language matcher here may be different than that for
// document due to available languages Wikipedia supports
type Wikipedia struct {
	language.Matcher
}

var (
	bufpool   *bpool.BufferPool // makes sure no errors when writing to our templates
	templates map[string]*template.Template
)

func init() {
	bufpool = bpool.NewBufferPool(48) // what is the appropriate size??? 48??? 64???
}

type response struct {
	status   int
	redirect string
	template string
	data     interface{}
	err      error
}

type appHandler func(http.ResponseWriter, *http.Request) *response

// middleware sets a timeout and then serves.
func (f *Frontend) middleware(next appHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rsp := fn(w, r); rsp != nil {
		switch rsp.status {
		case http.StatusOK:
			buf := bufpool.Get()
			defer bufpool.Put(buf)

			switch rsp.template {
			case "json":
				w.Header().Set("Content-Type", "application/json") // the default for json is utf-8
				err := json.NewEncoder(buf).Encode(rsp.data)
				if err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}
			default: // html by default
				w.Header().Set("Content-Type", "text/html; charset=utf-8")

				tmpl, ok := templates[rsp.template]
				if !ok {
					rsp.status = http.StatusInternalServerError
					rsp.err = fmt.Errorf("template doesn't exist: %q", rsp.template)
					errHandler(w, rsp)
					return
				}

				if err := tmpl.Execute(buf, rsp.data); err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}
			}

			if _, err := buf.WriteTo(w); err != nil {
				rsp.status, rsp.err = http.StatusInternalServerError, err
				errHandler(w, rsp)
			}
		case http.StatusFound: // !bang
			http.Redirect(w, r, rsp.redirect, http.StatusFound)
		case http.StatusBadRequest, http.StatusInternalServerError:
			errHandler(w, rsp)
		default:
			log.Info.Printf("Unknown status %d\n", rsp.status)
		}
	}
}

func errHandler(w http.ResponseWriter, rsp *response) {
	switch rsp.status {
	case http.StatusBadRequest:
		log.Debug.Println(rsp.err)
	case http.StatusInternalServerError:
		log.Info.Println(rsp.err)
	}

	http.Error(w, http.StatusText(rsp.status), rsp.status)
}

func (f *Frontend) autocompleteHandler(w http.ResponseWriter, r *http.Request) *response {
	var proxyFavIcon = func(u string) string {
		return fmt.Sprintf("/image/32x,s%v/%v", hmacKey(u), u)
	}

	q := strings.TrimSpace(r.FormValue("q"))

	if q == "!" {
		bngs := []bangs.Suggestion{}
		triggers := []string{"g", "a", "b", "reddit", "w"}
		for _, trigger := range triggers {
			for _, bng := range f.Bangs.Bangs {
				for _, tr := range bng.Triggers {
					if tr == trigger {
						sug := bangs.Suggestion{
							Trigger: trigger,
							Name:    bng.Name,
							FavIcon: proxyFavIcon(bng.FavIcon),
						}
						bngs = append(bngs, sug)
					}
				}
			}
		}

		// give a default set of !bang suggestions
		return &response{
			status:   http.StatusOK,
			template: "json",
			data: bangs.Results{
				Suggestions: bngs,
			},
		}

	} else if len(q) > 1 && !strings.HasPrefix(q, " ") && strings.HasPrefix(q, "!") {
		res, err := f.Bangs.Suggest(q, 10)
		if err != nil {
			return &response{
				status: http.StatusInternalServerError,
				err:    err,
			}
		}

		if len(res.Suggestions) > 0 {
			// fetch the favicons through our image proxy
			for i, b := range res.Suggestions {
				b.FavIcon = proxyFavIcon(b.FavIcon)
				res.Suggestions[i] = b
			}

			return &response{
				status:   http.StatusOK,
				template: "json",
				data:     res,
			}
		}
	}

	res, err := f.Suggest.Completion(q, 10)
	if err != nil {
		return &response{
			status: http.StatusInternalServerError,
			err:    err,
		}
	}

	return &response{
		status:   http.StatusOK,
		template: "json",
		data:     res,
	}
}

// ParseTemplates parses our html templates.
var ParseTemplates = func() {
	templates = make(map[string]*template.Template)
	templates["maps"] = template.Must(
		template.New("maps.html").
			Funcs(funcMap).
			ParseFiles(
				"templates/maps.html",
			),
	)
	templates["search"] = template.Must(
		template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"templates/base.html",
				"templates/search_form.html",
				"templates/search.html",
				"templates/wikipedia.html",
			),
	)
	templates["about"] = template.Must(
		template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"templates/base.html",
				"templates/search_form.html",
				"templates/about.html",
			),
	)
}
