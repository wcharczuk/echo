package webutil

import "net/url"

// MustParseURL parses a url and panics if there is an error.
func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
