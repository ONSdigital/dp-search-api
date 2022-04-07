// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	api "github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/clients"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/service"
	"net/http"
	"sync"
)

// Ensure, that InitialiserMock does implement service.Initialiser.
// If this is not the case, regenerate this file with moq.
var _ service.Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of service.Initialiser.
//
// 	func TestSomethingThatUsesInitialiser(t *testing.T) {
//
// 		// make and configure a mocked service.Initialiser
// 		mockedInitialiser := &InitialiserMock{
// 			DoGetAuthorisationHandlersFunc: func(cfg *config.Config) api.AuthHandler {
// 				panic("mock out the DoGetAuthorisationHandlers method")
// 			},
// 			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
// 				panic("mock out the DoGetHTTPServer method")
// 			},
// 			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 				panic("mock out the DoGetHealthCheck method")
// 			},
// 			DoGetHealthClientFunc: func(name string, url string) *health.Client {
// 				panic("mock out the DoGetHealthClient method")
// 			},
// 		}
//
// 		// use mockedInitialiser in code that requires service.Initialiser
// 		// and then make assertions.
//
// 	}
type InitialiserMock struct {
	// DoGetAuthorisationHandlersFunc mocks the DoGetAuthorisationHandlers method.
	DoGetAuthorisationHandlersFunc func(cfg *config.Config) api.AuthHandler

	// DoGetDatasetClientFunc mocks the DoGetDatasetClient method.
	DoGetDatasetClientFunc func(cfg *config.Config) clients.DatasetAPIClient

	// DoGetHTTPServerFunc mocks the DoGetHTTPServer method.
	DoGetHTTPServerFunc func(bindAddr string, router http.Handler) service.HTTPServer

	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error)

	// DoGetHealthClientFunc mocks the DoGetHealthClient method.
	DoGetHealthClientFunc func(name string, url string) *health.Client

	// calls tracks calls to the methods.
	calls struct {
		// DoGetAuthorisationHandlers holds details about calls to the DoGetAuthorisationHandlers method.
		DoGetAuthorisationHandlers []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetDatasetClient holds details about calls to the DoGetDatasetClient method.
		DoGetDatasetClient []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetHTTPServer holds details about calls to the DoGetHTTPServer method.
		DoGetHTTPServer []struct {
			// BindAddr is the bindAddr argument value.
			BindAddr string
			// Router is the router argument value.
			Router http.Handler
		}
		// DoGetHealthCheck holds details about calls to the DoGetHealthCheck method.
		DoGetHealthCheck []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// BuildTime is the buildTime argument value.
			BuildTime string
			// GitCommit is the gitCommit argument value.
			GitCommit string
			// Version is the version argument value.
			Version string
		}
		// DoGetHealthClient holds details about calls to the DoGetHealthClient method.
		DoGetHealthClient []struct {
			// Name is the name argument value.
			Name string
			// URL is the url argument value.
			URL string
		}
	}
	lockDoGetAuthorisationHandlers sync.RWMutex
	lockDoGetDatasetClient         sync.RWMutex
	lockDoGetHTTPServer            sync.RWMutex
	lockDoGetHealthCheck           sync.RWMutex
	lockDoGetHealthClient          sync.RWMutex
}

// DoGetAuthorisationHandlers calls DoGetAuthorisationHandlersFunc.
func (mock *InitialiserMock) DoGetAuthorisationHandlers(cfg *config.Config) api.AuthHandler {
	if mock.DoGetAuthorisationHandlersFunc == nil {
		panic("InitialiserMock.DoGetAuthorisationHandlersFunc: method is nil but Initialiser.DoGetAuthorisationHandlers was just called")
	}
	callInfo := struct {
		Cfg *config.Config
	}{
		Cfg: cfg,
	}
	mock.lockDoGetAuthorisationHandlers.Lock()
	mock.calls.DoGetAuthorisationHandlers = append(mock.calls.DoGetAuthorisationHandlers, callInfo)
	mock.lockDoGetAuthorisationHandlers.Unlock()
	return mock.DoGetAuthorisationHandlersFunc(cfg)
}

// DoGetAuthorisationHandlersCalls gets all the calls that were made to DoGetAuthorisationHandlers.
// Check the length with:
//     len(mockedInitialiser.DoGetAuthorisationHandlersCalls())
func (mock *InitialiserMock) DoGetAuthorisationHandlersCalls() []struct {
	Cfg *config.Config
} {
	var calls []struct {
		Cfg *config.Config
	}
	mock.lockDoGetAuthorisationHandlers.RLock()
	calls = mock.calls.DoGetAuthorisationHandlers
	mock.lockDoGetAuthorisationHandlers.RUnlock()
	mock.lockDoGetAuthorisationHandlers.RUnlock()
	return calls
}

// DoGetDatasetClient calls DoGetDatasetClientFunc.
func (mock *InitialiserMock) DoGetDatasetClient(cfg *config.Config) clients.DatasetAPIClient {
	if mock.DoGetDatasetClientFunc == nil {
		panic("InitialiserMock.DoGetDatasetClientFunc: method is nil but Initialiser.DoGetDatasetClient was just called")
	}
	callInfo := struct {
		Cfg *config.Config
	}{
		Cfg: cfg,
	}
	mock.lockDoGetDatasetClient.Lock()
	mock.calls.DoGetDatasetClient = append(mock.calls.DoGetDatasetClient, callInfo)
	mock.lockDoGetDatasetClient.Unlock()
	return mock.DoGetDatasetClientFunc(cfg)
}

// DoGetDatasetClientCalls gets all the calls that were made to DoGetDatasetClient.
// Check the length with:
//     len(mockedInitialiser.DoGetDatasetClientCalls())
func (mock *InitialiserMock) DoGetDatasetClientCalls() []struct {
	Cfg *config.Config
} {
	var calls []struct {
		Cfg *config.Config
	}
	mock.lockDoGetDatasetClient.RLock()
	calls = mock.calls.DoGetDatasetClient
	mock.lockDoGetDatasetClient.RUnlock()
	return calls
}

// DoGetHTTPServer calls DoGetHTTPServerFunc.
func (mock *InitialiserMock) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	if mock.DoGetHTTPServerFunc == nil {
		panic("InitialiserMock.DoGetHTTPServerFunc: method is nil but Initialiser.DoGetHTTPServer was just called")
	}
	callInfo := struct {
		BindAddr string
		Router   http.Handler
	}{
		BindAddr: bindAddr,
		Router:   router,
	}
	mock.lockDoGetHTTPServer.Lock()
	mock.calls.DoGetHTTPServer = append(mock.calls.DoGetHTTPServer, callInfo)
	mock.lockDoGetHTTPServer.Unlock()
	return mock.DoGetHTTPServerFunc(bindAddr, router)
}

// DoGetHTTPServerCalls gets all the calls that were made to DoGetHTTPServer.
// Check the length with:
//     len(mockedInitialiser.DoGetHTTPServerCalls())
func (mock *InitialiserMock) DoGetHTTPServerCalls() []struct {
	BindAddr string
	Router   http.Handler
} {
	var calls []struct {
		BindAddr string
		Router   http.Handler
	}
	mock.lockDoGetHTTPServer.RLock()
	calls = mock.calls.DoGetHTTPServer
	mock.lockDoGetHTTPServer.RUnlock()
	return calls
}

// DoGetHealthCheck calls DoGetHealthCheckFunc.
func (mock *InitialiserMock) DoGetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	if mock.DoGetHealthCheckFunc == nil {
		panic("InitialiserMock.DoGetHealthCheckFunc: method is nil but Initialiser.DoGetHealthCheck was just called")
	}
	callInfo := struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}{
		Cfg:       cfg,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		Version:   version,
	}
	mock.lockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	mock.lockDoGetHealthCheck.Unlock()
	return mock.DoGetHealthCheckFunc(cfg, buildTime, gitCommit, version)
}

// DoGetHealthCheckCalls gets all the calls that were made to DoGetHealthCheck.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthCheckCalls())
func (mock *InitialiserMock) DoGetHealthCheckCalls() []struct {
	Cfg       *config.Config
	BuildTime string
	GitCommit string
	Version   string
} {
	var calls []struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}
	mock.lockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	mock.lockDoGetHealthCheck.RUnlock()
	return calls
}

// DoGetHealthClient calls DoGetHealthClientFunc.
func (mock *InitialiserMock) DoGetHealthClient(name string, url string) *health.Client {
	if mock.DoGetHealthClientFunc == nil {
		panic("InitialiserMock.DoGetHealthClientFunc: method is nil but Initialiser.DoGetHealthClient was just called")
	}
	callInfo := struct {
		Name string
		URL  string
	}{
		Name: name,
		URL:  url,
	}
	mock.lockDoGetHealthClient.Lock()
	mock.calls.DoGetHealthClient = append(mock.calls.DoGetHealthClient, callInfo)
	mock.lockDoGetHealthClient.Unlock()
	return mock.DoGetHealthClientFunc(name, url)
}

// DoGetHealthClientCalls gets all the calls that were made to DoGetHealthClient.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthClientCalls())
func (mock *InitialiserMock) DoGetHealthClientCalls() []struct {
	Name string
	URL  string
} {
	var calls []struct {
		Name string
		URL  string
	}
	mock.lockDoGetHealthClient.RLock()
	calls = mock.calls.DoGetHealthClient
	mock.lockDoGetHealthClient.RUnlock()
	return calls
}
