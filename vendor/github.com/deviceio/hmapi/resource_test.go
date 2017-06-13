package hmapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"net/url"

	"net"

	"strings"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Test_ResourceRequest_when_calling_get struct {
	suite.Suite
}

func (t *Test_ResourceRequest_when_calling_get) Test_request_successfully_canceled() {
	objects := t.getTestServerAndClient()
	defer objects.Server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	objects.Mux.HandleFunc("/whatever", func(rw http.ResponseWriter, r *http.Request) {
		cancel() // we cancel the ctx here to simulate a request already in flight
		rw.WriteHeader(http.StatusOK)
	})

	res, err := objects.Client.Resource("/whatever").Get(ctx)

	assert.Nil(t.T(), res)
	assert.NotNil(t.T(), err)
	assert.True(t.T(), strings.Contains(err.Error(), "context canceled"))
}

func (t *Test_ResourceRequest_when_calling_get) Test_returns_error_when_non_http_200_response() {
	ret := t.getTestServerAndClient()
	defer ret.Server.Close()

	ret.Mux.HandleFunc("/resource", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	})

	resource, err := ret.Client.Resource("/resource").Get(context.Background())

	assert.Nil(t.T(), resource)
	assert.NotNil(t.T(), err)

	e, ok := err.(*ErrUnexpectedHTTPResponseStatus)
	assert.True(t.T(), ok)
	assert.Equal(t.T(), http.StatusOK, e.ExpectedStatus)
	assert.Equal(t.T(), http.StatusInternalServerError, e.ActualStatus)
}

func (t *Test_ResourceRequest_when_calling_get) Test_returns_error_when_failed_json_decode() {
	ret := t.getTestServerAndClient()
	defer ret.Server.Close()

	ret.Mux.HandleFunc("/resource", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("{some invalid json"))
	})

	resource, err := ret.Client.Resource("/resource").Get(context.Background())

	assert.Nil(t.T(), resource)
	assert.NotNil(t.T(), err)

	_, ok := err.(*ErrResourceUnmarshalFailure)
	assert.True(t.T(), ok)
}

func (t *Test_ResourceRequest_when_calling_get) Test_returns_minimal_resource() {
	ret := t.getTestServerAndClient()
	defer ret.Server.Close()

	ret.Mux.HandleFunc("/resource", func(rw http.ResponseWriter, r *http.Request) {
		resource := &Resource{
			Forms:   map[string]*Form{},
			Links:   map[string]*Link{},
			Content: map[string]*Content{},
		}

		json.NewEncoder(rw).Encode(resource)
	})

	resource, err := ret.Client.Resource("/resource").Get(context.Background())

	assert.NotNil(t.T(), resource)
	assert.Nil(t.T(), err)

	assert.NotNil(t.T(), resource.Content)
	assert.NotNil(t.T(), resource.Forms)
	assert.NotNil(t.T(), resource.Links)
}

func (t *Test_ResourceRequest_when_calling_get) getTestServerAndClient() (ret struct {
	Mux    *mux.Router
	Host   string
	Port   int
	Server *httptest.Server
	Client Client
}) {
	mux := mux.NewRouter()
	svr := httptest.NewServer(mux)

	url, _ := url.Parse(svr.URL)

	hoststr, portstr, _ := net.SplitHostPort(url.Host)
	port, _ := strconv.ParseInt(portstr, 10, 0)

	client := NewClient(&ClientConfig{
		Auth:   &AuthNone{},
		Host:   hoststr,
		Port:   int(port),
		Scheme: HTTP,
	})

	ret.Mux = mux
	ret.Host = hoststr
	ret.Port = int(port)
	ret.Server = svr
	ret.Client = client
	return
}

func TestResourceTestSuite(t *testing.T) {
	suite.Run(t, new(Test_ResourceRequest_when_calling_get))
}
