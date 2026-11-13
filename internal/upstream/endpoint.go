package upstream

import "net/url"

type Endpoint struct {
	ID  string
	URL *url.URL
}
