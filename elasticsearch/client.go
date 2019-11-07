package elasticsearch

// Client provides methods to wrap around the elastcsearch package in order to facilitate unit testing
type Client struct {
	URL string
}

// New creates a new elasticsearch client
func New(URL string) *Client {

	Setup(URL)

	return &Client{
		URL: URL,
	}
}

// MultiSearch is a method that wraps the MultiSerch function of the elasticsearch package
func (cli *Client) MultiSearch(index string, docType string, request []byte) ([]byte, error) {
	return MultiSearch(index, docType, request)
}
