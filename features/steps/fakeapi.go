package steps

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/maxcnunes/httpfake"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP                     *httpfake.HTTPFake
	outboundRequests             []string
	collectOutboundRequestBodies httpfake.CustomAssertor
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI(t testing.TB) *FakeAPI {
	fa := &FakeAPI{
		fakeHTTP: httpfake.New(httpfake.WithTesting(t)),
	}

	fa.collectOutboundRequestBodies = func(r *http.Request) error {
		// inspect request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading the outbound request body: %s", err.Error())
		}
		fa.outboundRequests = append(fa.outboundRequests, string(body))
		return nil
	}

	return fa
}

func (f *FakeAPI) setJSONResponseForPost(url string, statusCode int, body []byte) {
	f.fakeHTTP.NewHandler().Post(url).Reply(statusCode).Body(body)
}

func (f *FakeAPI) setJSONResponseForGetHealth(url string, statusCode int, body []byte) {
	f.fakeHTTP.NewHandler().Post(url).Reply(statusCode).Body(body)
}

// Close closes the fake API
func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}

// Reset resets the fake API
func (f *FakeAPI) Reset() {
	f.fakeHTTP.Reset()
}
