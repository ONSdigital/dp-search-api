// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/sdk"
	apiError "github.com/ONSdigital/dp-search-api/sdk/errors"
	"github.com/ONSdigital/dp-search-api/transformer"
	"sync"
)

// Ensure, that ClienterMock does implement sdk.Clienter.
// If this is not the case, regenerate this file with moq.
var _ sdk.Clienter = &ClienterMock{}

// ClienterMock is a mock implementation of sdk.Clienter.
//
//	func TestSomethingThatUsesClienter(t *testing.T) {
//
//		// make and configure a mocked sdk.Clienter
//		mockedClienter := &ClienterMock{
//			CheckerFunc: func(ctx context.Context, check *health.CheckState) error {
//				panic("mock out the Checker method")
//			},
//			CreateIndexFunc: func(ctx context.Context, options sdk.Options) (*models.CreateIndexResponse, apiError.Error) {
//				panic("mock out the CreateIndex method")
//			},
//			GetReleaseCalendarEntriesFunc: func(ctx context.Context, options sdk.Options) (*transformer.SearchReleaseResponse, apiError.Error) {
//				panic("mock out the GetReleaseCalendarEntries method")
//			},
//			GetSearchFunc: func(ctx context.Context, options sdk.Options) (*models.SearchResponse, apiError.Error) {
//				panic("mock out the GetSearch method")
//			},
//			HealthFunc: func() *healthcheck.Client {
//				panic("mock out the Health method")
//			},
//			PostSearchURIsFunc: func(ctx context.Context, options sdk.Options, urisRequest api.URIsRequest) (*models.SearchResponse, apiError.Error) {
//				panic("mock out the PostSearchURIs method")
//			},
//			URLFunc: func() string {
//				panic("mock out the URL method")
//			},
//		}
//
//		// use mockedClienter in code that requires sdk.Clienter
//		// and then make assertions.
//
//	}
type ClienterMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, check *health.CheckState) error

	// CreateIndexFunc mocks the CreateIndex method.
	CreateIndexFunc func(ctx context.Context, options sdk.Options) (*models.CreateIndexResponse, apiError.Error)

	// GetReleaseCalendarEntriesFunc mocks the GetReleaseCalendarEntries method.
	GetReleaseCalendarEntriesFunc func(ctx context.Context, options sdk.Options) (*transformer.SearchReleaseResponse, apiError.Error)

	// GetSearchFunc mocks the GetSearch method.
	GetSearchFunc func(ctx context.Context, options sdk.Options) (*models.SearchResponse, apiError.Error)

	// HealthFunc mocks the Health method.
	HealthFunc func() *healthcheck.Client

	// PostSearchURIsFunc mocks the PostSearchURIs method.
	PostSearchURIsFunc func(ctx context.Context, options sdk.Options, urisRequest api.URIsRequest) (*models.SearchResponse, apiError.Error)

	// URLFunc mocks the URL method.
	URLFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Check is the check argument value.
			Check *health.CheckState
		}
		// CreateIndex holds details about calls to the CreateIndex method.
		CreateIndex []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Options is the options argument value.
			Options sdk.Options
		}
		// GetReleaseCalendarEntries holds details about calls to the GetReleaseCalendarEntries method.
		GetReleaseCalendarEntries []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Options is the options argument value.
			Options sdk.Options
		}
		// GetSearch holds details about calls to the GetSearch method.
		GetSearch []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Options is the options argument value.
			Options sdk.Options
		}
		// Health holds details about calls to the Health method.
		Health []struct {
		}
		// PostSearchURIs holds details about calls to the PostSearchURIs method.
		PostSearchURIs []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Options is the options argument value.
			Options sdk.Options
			// UrisRequest is the urisRequest argument value.
			UrisRequest api.URIsRequest
		}
		// URL holds details about calls to the URL method.
		URL []struct {
		}
	}
	lockChecker                   sync.RWMutex
	lockCreateIndex               sync.RWMutex
	lockGetReleaseCalendarEntries sync.RWMutex
	lockGetSearch                 sync.RWMutex
	lockHealth                    sync.RWMutex
	lockPostSearchURIs            sync.RWMutex
	lockURL                       sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *ClienterMock) Checker(ctx context.Context, check *health.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("ClienterMock.CheckerFunc: method is nil but Clienter.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Check *health.CheckState
	}{
		Ctx:   ctx,
		Check: check,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, check)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//
//	len(mockedClienter.CheckerCalls())
func (mock *ClienterMock) CheckerCalls() []struct {
	Ctx   context.Context
	Check *health.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		Check *health.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// CreateIndex calls CreateIndexFunc.
func (mock *ClienterMock) CreateIndex(ctx context.Context, options sdk.Options) (*models.CreateIndexResponse, apiError.Error) {
	if mock.CreateIndexFunc == nil {
		panic("ClienterMock.CreateIndexFunc: method is nil but Clienter.CreateIndex was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Options sdk.Options
	}{
		Ctx:     ctx,
		Options: options,
	}
	mock.lockCreateIndex.Lock()
	mock.calls.CreateIndex = append(mock.calls.CreateIndex, callInfo)
	mock.lockCreateIndex.Unlock()
	return mock.CreateIndexFunc(ctx, options)
}

// CreateIndexCalls gets all the calls that were made to CreateIndex.
// Check the length with:
//
//	len(mockedClienter.CreateIndexCalls())
func (mock *ClienterMock) CreateIndexCalls() []struct {
	Ctx     context.Context
	Options sdk.Options
} {
	var calls []struct {
		Ctx     context.Context
		Options sdk.Options
	}
	mock.lockCreateIndex.RLock()
	calls = mock.calls.CreateIndex
	mock.lockCreateIndex.RUnlock()
	return calls
}

// GetReleaseCalendarEntries calls GetReleaseCalendarEntriesFunc.
func (mock *ClienterMock) GetReleaseCalendarEntries(ctx context.Context, options sdk.Options) (*transformer.SearchReleaseResponse, apiError.Error) {
	if mock.GetReleaseCalendarEntriesFunc == nil {
		panic("ClienterMock.GetReleaseCalendarEntriesFunc: method is nil but Clienter.GetReleaseCalendarEntries was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Options sdk.Options
	}{
		Ctx:     ctx,
		Options: options,
	}
	mock.lockGetReleaseCalendarEntries.Lock()
	mock.calls.GetReleaseCalendarEntries = append(mock.calls.GetReleaseCalendarEntries, callInfo)
	mock.lockGetReleaseCalendarEntries.Unlock()
	return mock.GetReleaseCalendarEntriesFunc(ctx, options)
}

// GetReleaseCalendarEntriesCalls gets all the calls that were made to GetReleaseCalendarEntries.
// Check the length with:
//
//	len(mockedClienter.GetReleaseCalendarEntriesCalls())
func (mock *ClienterMock) GetReleaseCalendarEntriesCalls() []struct {
	Ctx     context.Context
	Options sdk.Options
} {
	var calls []struct {
		Ctx     context.Context
		Options sdk.Options
	}
	mock.lockGetReleaseCalendarEntries.RLock()
	calls = mock.calls.GetReleaseCalendarEntries
	mock.lockGetReleaseCalendarEntries.RUnlock()
	return calls
}

// GetSearch calls GetSearchFunc.
func (mock *ClienterMock) GetSearch(ctx context.Context, options sdk.Options) (*models.SearchResponse, apiError.Error) {
	if mock.GetSearchFunc == nil {
		panic("ClienterMock.GetSearchFunc: method is nil but Clienter.GetSearch was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Options sdk.Options
	}{
		Ctx:     ctx,
		Options: options,
	}
	mock.lockGetSearch.Lock()
	mock.calls.GetSearch = append(mock.calls.GetSearch, callInfo)
	mock.lockGetSearch.Unlock()
	return mock.GetSearchFunc(ctx, options)
}

// GetSearchCalls gets all the calls that were made to GetSearch.
// Check the length with:
//
//	len(mockedClienter.GetSearchCalls())
func (mock *ClienterMock) GetSearchCalls() []struct {
	Ctx     context.Context
	Options sdk.Options
} {
	var calls []struct {
		Ctx     context.Context
		Options sdk.Options
	}
	mock.lockGetSearch.RLock()
	calls = mock.calls.GetSearch
	mock.lockGetSearch.RUnlock()
	return calls
}

// Health calls HealthFunc.
func (mock *ClienterMock) Health() *healthcheck.Client {
	if mock.HealthFunc == nil {
		panic("ClienterMock.HealthFunc: method is nil but Clienter.Health was just called")
	}
	callInfo := struct {
	}{}
	mock.lockHealth.Lock()
	mock.calls.Health = append(mock.calls.Health, callInfo)
	mock.lockHealth.Unlock()
	return mock.HealthFunc()
}

// HealthCalls gets all the calls that were made to Health.
// Check the length with:
//
//	len(mockedClienter.HealthCalls())
func (mock *ClienterMock) HealthCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockHealth.RLock()
	calls = mock.calls.Health
	mock.lockHealth.RUnlock()
	return calls
}

// PostSearchURIs calls PostSearchURIsFunc.
func (mock *ClienterMock) PostSearchURIs(ctx context.Context, options sdk.Options, urisRequest api.URIsRequest) (*models.SearchResponse, apiError.Error) {
	if mock.PostSearchURIsFunc == nil {
		panic("ClienterMock.PostSearchURIsFunc: method is nil but Clienter.PostSearchURIs was just called")
	}
	callInfo := struct {
		Ctx         context.Context
		Options     sdk.Options
		UrisRequest api.URIsRequest
	}{
		Ctx:         ctx,
		Options:     options,
		UrisRequest: urisRequest,
	}
	mock.lockPostSearchURIs.Lock()
	mock.calls.PostSearchURIs = append(mock.calls.PostSearchURIs, callInfo)
	mock.lockPostSearchURIs.Unlock()
	return mock.PostSearchURIsFunc(ctx, options, urisRequest)
}

// PostSearchURIsCalls gets all the calls that were made to PostSearchURIs.
// Check the length with:
//
//	len(mockedClienter.PostSearchURIsCalls())
func (mock *ClienterMock) PostSearchURIsCalls() []struct {
	Ctx         context.Context
	Options     sdk.Options
	UrisRequest api.URIsRequest
} {
	var calls []struct {
		Ctx         context.Context
		Options     sdk.Options
		UrisRequest api.URIsRequest
	}
	mock.lockPostSearchURIs.RLock()
	calls = mock.calls.PostSearchURIs
	mock.lockPostSearchURIs.RUnlock()
	return calls
}

// URL calls URLFunc.
func (mock *ClienterMock) URL() string {
	if mock.URLFunc == nil {
		panic("ClienterMock.URLFunc: method is nil but Clienter.URL was just called")
	}
	callInfo := struct {
	}{}
	mock.lockURL.Lock()
	mock.calls.URL = append(mock.calls.URL, callInfo)
	mock.lockURL.Unlock()
	return mock.URLFunc()
}

// URLCalls gets all the calls that were made to URL.
// Check the length with:
//
//	len(mockedClienter.URLCalls())
func (mock *ClienterMock) URLCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockURL.RLock()
	calls = mock.calls.URL
	mock.lockURL.RUnlock()
	return calls
}
