package frontend

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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

	css := r.FormValue("css")
	if css != "" {
		u, err := url.Parse(css)
		if err != nil {
			log.Info.Println(err)
		}

		res, err := f.get(u)
		if err != nil {
			log.Info.Println(err)
		}

		defer res.Body.Close()

		h, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Info.Println(err)
		}

		resp.data = replaceCSS(u, string(h))

		resp.template = "proxy_css"
		return resp
	}

	u := r.FormValue("u")
	if u == "" {
		return resp
	}

	base, err := url.Parse(u)
	if err != nil {
		log.Info.Println(err)
	}

	res, err := f.get(base)
	if err != nil {
		log.Info.Println(err)
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Info.Println(err)
	}

	// TODO: remove all comments...no need for them

	// remove all javascript
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// disable all forms
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("disabled", "disabled") // doesn't seem to work....
		s.SetAttr("action", "javascript:void(0);")
	})

	// proxy links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		for _, href := range []string{"href"} {
			if lnk, ok := s.Attr(href); ok {
				u, err := createProxyLink(base, lnk)
				if err != nil {
					log.Info.Println(err)
				}

				s.SetAttr(href, u.String())
				s.SetAttr("target", "_top") // make all links open in the main page (not w/in the iframe)
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

				matches := reSrcSet.FindAllStringSubmatch(lnk, -1)
				if len(matches) == 0 {
					u := createProxyImage(base, lnk)
					s.SetAttr(src, u)
					continue
				}

				lnks := []string{}

				for _, m := range matches {
					if isBase64(m[1]) {
						lnks = append(lnks, m[1])
						continue
					}

					u := createProxyImage(base, m[1])

					lnks = append(lnks, fmt.Sprintf("%v %v", u, m[2]))
				}

				s.SetAttr(src, strings.Join(lnks, " "))
			}
		}
	})

	// proxy url() within style tags
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		h := replaceCSS(base, s.Text())
		s.ReplaceWithHtml(fmt.Sprintf(`<style>%v</style>`, h))
	})

	// replace external stylesheets with a proxied version
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		if rel, ok := s.Attr("rel"); ok && strings.ToLower(rel) == "stylesheet" {
			if lnk, ok := s.Attr("href"); ok {
				u, err := createProxyCSSLink(base, lnk)
				if err != nil {
					log.Info.Println(err)
				}

				s.SetAttr("href", u.String())
			}
		}
	})

	h, err := doc.Html()
	if err != nil {
		log.Info.Println(err)
	}

	resp.data = proxyResponse{
		Brand: f.Brand,
		HTML:  h,
	}

	return resp
}

func isBase64(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "data:")
}

// can have ', ", or no quotes
var reCSSLinkReplacer = regexp.MustCompile(`(url\(['"]?)(?P<link>.*?)['"]?\)`)
var reSrcSet = regexp.MustCompile(`(?P<url>.*?),? (?P<size>[0-9]+(\.[0-9]+)?[wx],?)?\s?`)

// https://stackoverflow.com/a/28005189/522962
func replaceAllSubmatchFunc(re *regexp.Regexp, b []byte, f func(s []byte) []byte) []byte {
	idxs := re.FindAllSubmatchIndex(b, -1)
	if len(idxs) == 0 {
		return b
	}
	l := len(idxs)
	ret := append([]byte{}, b[:idxs[0][0]]...)
	for i, pair := range idxs {
		ret = append(ret, f(b[pair[4]:pair[5]])...) // 2 & 3 are <url>. 4 & 5 are the <link>
		if i+1 < l {
			ret = append(ret, b[pair[1]:idxs[i+1][0]]...)
		}
	}
	ret = append(ret, b[idxs[len(idxs)-1][1]:]...)
	return ret
}

func replaceCSS(base *url.URL, s string) string {
	// replace any urls with a proxied link
	ss := replaceAllSubmatchFunc(reCSSLinkReplacer, []byte(s), func(ss []byte) []byte {
		if isBase64(string(ss)) { // base64 image
			return []byte(fmt.Sprintf("url(%q)", ss))
		}

		m := string(ss)

		u, err := url.Parse(m)
		if err != nil {
			log.Info.Println(err)
		}

		u = base.ResolveReference(u)

		key := hmacKey(u.String())
		uu := fmt.Sprintf("/image/,s%v/%v", key, u.String())
		return []byte(fmt.Sprintf("url(%q)", uu))
	})

	return string(ss)
}

func createProxyImage(base *url.URL, lnk string) string {
	u, err := url.Parse(lnk)
	if err != nil {
		panic(err)
	}

	u = base.ResolveReference(u)
	key := hmacKey(u.String())
	l := fmt.Sprintf("/image/,s%v/%v", key, u.String())
	return l
}

func createProxyCSSLink(base *url.URL, lnk string) (*url.URL, error) {
	u, err := url.Parse(lnk)
	if err != nil {
		panic(err)
	}

	u = base.ResolveReference(u)

	uu, err := url.Parse("/proxy")
	if err != nil {
		return nil, err
	}

	q := uu.Query()
	q.Add("key", hmacKey(u.String()))
	q.Add("css", u.String())
	uu.RawQuery = q.Encode()
	return uu, err
}

func createProxyLink(base *url.URL, lnk string) (*url.URL, error) {
	u, err := url.Parse(lnk)
	if err != nil {
		panic(err)
	}

	u = base.ResolveReference(u)

	uu, err := url.Parse("/proxy")
	if err != nil {
		return nil, err
	}

	q := uu.Query()
	q.Add("key", hmacKey(u.String()))
	q.Add("u", u.String())
	uu.RawQuery = q.Encode()
	return uu, err
}

func (f *Frontend) get(u *url.URL) (*http.Response, error) {
	// we don't want &httputil.ReverseProxy as we don't want to pass the user's IP address & other info.
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	return f.ProxyClient.Do(request)
}
