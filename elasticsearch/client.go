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

// Search is a method that wraps the Search function of the elasticsearch package
func (cli *Client) Search(index string, docType string, request []byte) ([]byte, error) {
	return Search(index, docType, request)
}

// MultiSearch is a method that wraps the MultiSearch function of the elasticsearch package
func (cli *Client) MultiSearch(index string, docType string, request []byte) ([]byte, error) {
	return MultiSearch(index, docType, request)
}
