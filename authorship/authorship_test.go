package authorship

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

func TestParse(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/", http.StripPrefix("/test/", http.FileServer(http.Dir("testdata"))))

	s := httptest.NewServer(mux)
	defer s.Close()

	tcs := map[string]struct {
		HTML string
		URL  string
		Out  interface{}
	}{
		"No h-card": {
			HTML: `<p>nowt</p>`,
			Out:  nil,
		},
		// https://authorship.rocks/
		"Entry with p-author": {
			URL: "/test/1",
			Out: map[string][]interface{}{
				"name": {"William Shakespeare"},
			},
		},
		"Entry with h-card": {
			URL: "/test/2",
			Out: map[string][]interface{}{
				"name":  {"Homer"},
				"url":   {"https://en.wikiquote.org/wiki/Homer"},
				"photo": {s.URL + "/images/homer.jpg"},
			},
		},
		"Entry with separate h-card and rel=author": {
			URL: "/test/3",
			Out: map[string][]interface{}{
				"name": {"Patañjali"},
				"url":  {s.URL + "/test/3/about-patanjali"},
			},
		},
		"Entry with rel=author to an h-card with rel=me": {
			URL: "/test/4",
			Out: map[string][]interface{}{
				"name":  {"Virginia Woolf"},
				"url":   {s.URL + "/test/4/about-virginia-woolf"},
				"photo": {s.URL + "/images/virginia-woolf.jpg"},
			},
		},
		"Entry with rel=author to an h-card with u-url and u-uid": {
			URL: "/test/5",
			Out: map[string][]interface{}{
				"name":  {"Basho"},
				"url":   {s.URL + "/test/5/about-basho"},
				"uid":   {s.URL + "/test/5/about-basho"},
				"photo": {s.URL + "/images/basho.jpg"},
			},
		},
	}

	sURL, _ := url.Parse(s.URL)

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			var r io.Reader = strings.NewReader(tc.HTML)
			if tc.URL != "" {
				resp, _ := http.Get(s.URL + tc.URL)
				r = resp.Body
				defer resp.Body.Close()
			}

			author := Parse(r, sURL)

			assert.Equal(t, tc.Out, author)
		})
	}
}
