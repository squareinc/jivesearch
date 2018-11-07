package frontend

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jivesearch/jivesearch/log"
)

type proxyResponse struct {
	Brand
	Context `json:"-"`
	HTML    string `json:"-"`
}

func (f *Frontend) proxyHeaderHandler(w http.ResponseWriter, r *http.Request) *response {
	resp := &response{
		status:   http.StatusOK,
		template: "proxy_header",
		err:      nil,
	}

	resp.data = proxyResponse{
		Brand: f.Brand,
	}

	return resp
}

func (f *Frontend) proxyHandler(w http.ResponseWriter, r *http.Request) *response {
	resp := &response{
		status:   http.StatusOK,
		template: "proxy",
		data: proxyResponse{
			Brand: f.Brand,
		},
		err: nil,
	}

	u := r.FormValue("u")
	if u == "" {
		return resp
	}

	base, err := url.Parse(u)
	if err != nil {
		log.Info.Println(err)
	}

	fmt.Println(base)

	res, err := get(base.String())
	if err != nil {
		log.Info.Println(err)
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Info.Println(err)
	}

	fmt.Println("Yo, we are removing all <li> tags....remember to remove this part...")
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// TODO: remove all comments...no need for them

	// remove all javascript
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// disable all forms
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("disabled", "disabled")
	})

	// proxy links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		for _, href := range []string{"href"} {
			if lnk, ok := s.Attr(href); ok {
				u, err := url.Parse(lnk)
				if err != nil {
					log.Info.Println(err)
				}

				u, err = createProxyLink(base.ResolveReference(u))
				if err != nil {
					log.Info.Println(err)
				}

				s.SetAttr(href, u.String())
			}
		}
	})

	// proxy images
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		for _, src := range []string{"src", "srcset"} {
			if lnk, ok := s.Attr(src); ok {
				if lnk == "" {
					continue
				}

				if src == "srcset" {
					lnk = strings.Fields(lnk)[0]
				}

				u, err := url.Parse(lnk)
				if err != nil {
					log.Info.Println(err)
				}

				u = base.ResolveReference(u)
				key := hmacKey(u.String())
				l := fmt.Sprintf("/image/,s%v/%v", key, u.String())
				s.SetAttr(src, l)
			}
		}
	})

	// proxy url() within style tags
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		h := replaceCSS(s.Text())

		// replace the link with the css
		s.ReplaceWithHtml(fmt.Sprintf(`<style>%v</style>`, h))
	})

	// within each external css file, proxy all url() items
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		if rel, ok := s.Attr("rel"); ok && strings.ToLower(rel) == "stylesheet" {
			if lnk, ok := s.Attr("href"); ok {
				u, err := url.Parse(lnk)
				if err != nil {
					log.Info.Println(err)
				}

				u = base.ResolveReference(u)
				res, err := get(u.String())
				if err != nil {
					log.Info.Println(err)
				}

				defer res.Body.Close()

				h, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Info.Println(err)
				}

				st := replaceCSS(string(h))

				// replace the link with the css
				s.ReplaceWithHtml(fmt.Sprintf(`<style>%v</style>`, st))
			}
		} else { // just proxy the href
			if lnk, ok := s.Attr("href"); ok {
				u, err := url.Parse(lnk)
				if err != nil {
					log.Info.Println(err)
				}

				u, err = createProxyLink(base.ResolveReference(u))
				if err != nil {
					log.Info.Println(err)
				}

				s.SetAttr("href", u.String())
			}
		}
	})

	h, err := doc.Html()
	//_, err = doc.Html()
	if err != nil {
		log.Info.Println(err)
	}
	//fmt.Println(h)

	resp.data = proxyResponse{
		Brand: f.Brand,
		HTML:  h,
	}

	return resp
}

// can have ', ", or no quotes
var reCSSLinkReplacer = regexp.MustCompile(`url\(['"]?(?P<url>.*?)['"]?\)`)

func replaceCSS(s string) string {
	// replace any urls with a proxied link
	s = reCSSLinkReplacer.ReplaceAllStringFunc(s, func(m string) string {
		key := hmacKey(m)
		u := fmt.Sprintf("/image/,s%v/%v", key, m)
		return fmt.Sprintf("url(/proxy/%v)", u)
	})

	return s
}

func createProxyLink(u *url.URL) (*url.URL, error) {
	uu, err := url.Parse("/proxy")
	if err != nil {
		return nil, err
	}

	q := uu.Query()
	q.Add("key", hmacKey(u.String()))
	q.Add("url", u.String())
	uu.RawQuery = q.Encode()
	return uu, err
}

func get(u string) (*http.Response, error) {
	// we don't want &httputil.ReverseProxy as we don't want to pass the user's IP address & other info.
	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	request, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return client.Do(request)
}
