package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/jarcoal/httpmock"
	"github.com/jivesearch/jivesearch/log"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

func TestProxyHeaderHandler(t *testing.T) {
	for _, c := range []struct {
		name string
		u    string
		want *response
	}{
		{
			"basic", "https://example.com",
			&response{
				status:   http.StatusOK,
				template: "proxy_header",
				data: proxyResponse{
					Brand: Brand{
						Name:    "Some Name",
						TagLine: "A great tagline",
					},
					URL: "https://example.com",
				},
				err: nil,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{
				Brand: Brand{
					Name:    "Some Name",
					TagLine: "A great tagline",
				},
			}

			req, err := http.NewRequest("GET", "/proxy_header", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("u", c.u)
			req.URL.RawQuery = q.Encode()

			got := f.proxyHeaderHandler(httptest.NewRecorder(), req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestProxyHandler(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		css    string
		u      string
		key    string
		secret string
		html   string
	}

	for _, c := range []struct {
		name string
		args
		want string
	}{
		{
			"basic",
			args{
				css:    "",
				u:      "https://example.com",
				key:    "jfsdijf89sd",
				secret: "my_secret",
				html: `<html>
								<head>
									<script>alert("this is dangerous!")</script>
								</head>
								<body>
									<form id="form">
									</form>
									<a href="https://www.example.com">A link</a>
									<a href="/relative/link">A relative link</a>
									<iframe src="https://example.com/iframe/stuff"></iframe>
								</body>
							</html>`,
			},
			`<html>
				<head></head>
				<body>
					<form id=form disabled action=javascript:void(0);></form>
					<a href="/proxy?key=_Zbla8JTucVtfb7n-QIGsrKozkTGaGsuKlxppnXb6xM%3D&amp;u=https%3A%2F%2Fwww.example.com" target=_top>A link</a>
					<a href="/proxy?key=j_gIsLDElFG1Qnp3TAYn1KD5dwvJ0gB_KqvUjXvM64g%3D&amp;u=https%3A%2F%2Fexample.com%2Frelative%2Flink" target=_top>A relative link</a>
					<iframe src="/proxy?iframe=true&amp;key=QtzD41Rkf5VUsmVPv9kSn4VHfUqf2jMljGktkjYVOVc%3D&amp;u=https%3A%2F%2Fexample.com%2Fiframe%2Fstuff"></iframe>
				</body>
			</html>`,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			log.Debug.SetOutput(os.Stdout)

			f := &Frontend{
				Brand:       Brand{},
				ProxyClient: &http.Client{},
			}

			responder := httpmock.NewStringResponder(200, c.html)
			httpmock.RegisterResponder("GET", c.args.u, responder)

			req, err := http.NewRequest("GET", "/proxy", nil)
			if err != nil {
				t.Fatal(err)
			}

			hmacSecret = func() string { return c.args.secret }
			k := hmacKey(c.u)

			q := req.URL.Query()
			q.Add("css", c.css)
			q.Add("u", c.u)
			q.Add("key", k)
			req.URL.RawQuery = q.Encode()

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(c.want))
			if err != nil {
				t.Fatal(err)
			}

			s, err := doc.Html()
			if err != nil {
				t.Fatal(err)
			}

			s, err = htmlMinify(s)
			if err != nil {
				t.Fatal(err)
			}

			want := &response{
				status:   http.StatusOK,
				template: "proxy",
				data: proxyResponse{
					Brand: Brand{},
					HTML:  s,
					URL:   c.args.u,
				},
			}

			got := f.proxyHandler(httptest.NewRecorder(), req)
			g := got.data.(proxyResponse)
			g.HTML, err = htmlMinify(g.HTML)
			if err != nil {
				t.Fatal(err)
			}

			got.data = g

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %+v; want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}

func htmlMinify(s string) (string, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	return m.String("text/html", s)
}
