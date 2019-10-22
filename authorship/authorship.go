// An implementation of the indieweb authorship algorithm.
//
// Documented at https://indieweb.org/authorship
package authorship

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	mf2 "github.com/andyleap/microformats"
)

// Parse r adding the author to all hentries found.
func Parse(r io.Reader, baseURL *url.URL) *mf2.Data {
	data := mf2.New().Parse(r, baseURL)

	return Find(data, baseURL)
}

// Find the author for each hentry, replacing the author property with the
// determined properties.
func Find(data *mf2.Data, baseURL *url.URL) *mf2.Data {
	for _, item := range data.Items {
		if contains("h-entry", item.Type) {
			author := FindForHEntry(data, item, baseURL)
			if author != nil {
				item.Properties["author"] = []interface{}{author}
			}
		}
	}

	return data
}

// Find the author for the given hentry, replacing the author property with the
// determined properties.
func FindForHEntry(data *mf2.Data, hentry *mf2.MicroFormat, baseURL *url.URL) interface{} {
	// 1. start with a particular h-entry to determine authorship for, and no
	//    author. if no h-entry, then there's no post to find authorship for,
	//    abort.
	//
	// 2. parse the h-entry

	if hentry == nil {
		return nil
	}

	var author interface{}

	// 3. if the h-entry has an author property, use that

	if anAuthor, ok := hentry.Properties["author"]; ok && len(anAuthor) > 0 {
		author = anAuthor[0]
	} else if anAuthor, ok := data.Rels["author"]; ok && len(anAuthor) > 0 {
		author = anAuthor[0]
	} else {
		// 4. otherwise if the h-entry has a parent h-feed with author property, use
		//    that
		var hfeed *mf2.MicroFormat

		for _, item := range data.Items {
			if contains("h-feed", item.Type) {
				hfeed = item
				break
			}
		}

		if hfeed != nil {
			if anAuthor, ok := hfeed.Properties["author"]; ok && len(anAuthor) > 0 {
				author = anAuthor[0]
			}
		}
	}

	var authorPage string

	// 5. if an author property was found
	if author != nil {
		// a. if it has an h-card, use it, exit.
		if mfAuthor, ok := author.(*mf2.MicroFormat); ok && contains("h-card", mfAuthor.Type) {
			return mfAuthor.Properties
		}

		// b. otherwise if author property is an http(s) URL, let the author-page
		//    have that URL
		if sAuthor, ok := author.(string); ok && strings.HasPrefix(sAuthor, "http://") {
			authorPage = sAuthor
		} else {
			// c. otherwise use the author property as the author name, exit
			return map[string][]interface{}{
				"name": {author},
			}
		}
	}

	// 6. if there is no author-page and the h-entry's page is a permalink page,
	//    then
	if authorPage == "" && len(hentry.Properties["url"]) > 0 {
		if sURL, ok := hentry.Properties["url"][0].(string); ok && sURL == baseURL.String() {
			//   a. if the page has a rel-author link, let the author-page's URL be the
			//      href of the rel-author link
			if len(data.Rels["author"]) > 0 {
				authorPage = data.Rels["author"][0]
			}
		}
	}

	// 7. if there is an author-page URL
	if authorPage != "" {
		//   a. get the author-page from that URL and parse it for microformats2
		authorURL, err := url.Parse(authorPage)
		if err != nil {
			return nil
		}

		resp, err := http.Get(authorPage)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		authorData := mf2.New().Parse(resp.Body, authorURL)

		//   b. if author-page has 1+ h-card with url == uid == author-page's URL,
		//      then use first such h-card, exit.
		for _, item := range authorData.Items {
			if contains("h-card", item.Type) && len(item.Properties["url"]) > 0 && len(item.Properties["uid"]) > 0 {
				sURL, urlOK := item.Properties["url"][0].(string)
				sUID, uidOK := item.Properties["uid"][0].(string)

				if urlOK && uidOK && sURL == sUID && sURL == authorPage {
					return item.Properties
				}
			}
		}

		//   c. else if author-page has 1+ h-card with url property which matches the
		//      href of a rel-me link on the author-page (perhaps the same hyperlink
		//      element as the u-url, though not required to be), use first such
		//      h-card, exit.
		for _, item := range authorData.Items {
			if contains("h-card", item.Type) && len(item.Properties["url"]) > 0 {
				sURL, urlOK := item.Properties["url"][0].(string)

				if urlOK {
					for _, me := range authorData.Rels["me"] {
						if me == sURL {
							return item.Properties
						}
					}
				}
			}
		}

		//   d. if the h-entry's page has 1+ h-card with url == author-page URL, use
		//      first such h-card, exit.
		for _, item := range authorData.Items {
			if contains("h-card", item.Type) && len(item.Properties["url"]) > 0 {
				sURL, urlOK := item.Properties["url"][0].(string)

				if urlOK && sURL == authorPage {
					return item.Properties
				}
			}
		}
	}

	return nil
}

func contains(needle string, xs []string) bool {
	for _, x := range xs {
		if x == needle {
			return true
		}
	}

	return false
}
