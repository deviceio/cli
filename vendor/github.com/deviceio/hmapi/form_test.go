package hmapi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Test_FormRequest_when_calling_submit struct {
	suite.Suite
}

func (t *Test_FormRequest_when_calling_submit) Test_submission_canceled_successfully() {
	objects := t.getTestServerAndClient()
	defer objects.Server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	objects.Mux.HandleFunc("/resource/test", func(rw http.ResponseWriter, r *http.Request) {
		cancel() // we cancel the ctx here to simulate a request already in flight
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("response"))
	}).Methods("POST")

	objects.Mux.HandleFunc("/resource", func(rw http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(&Resource{
			Forms: map[string]*Form{
				"test": &Form{
					Action:  "/resource/test",
					Method:  POST,
					Enctype: MediaTypeMultipartFormData,
					Type:    "none",
					Fields: []*FormField{
						&FormField{
							Name:     "foo",
							Type:     MediaTypeHMAPIString,
							Required: true,
						},
					},
				},
			},
		})

		if err != nil {
			assert.FailNow(t.T(), "error marshal json", err)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}).Methods("GET")

	resp, err := objects.Client.Resource("/resource").Form("test").AddFieldAsString("foo", "test").Submit(ctx)

	assert.Nil(t.T(), resp)
	assert.NotNil(t.T(), err)
	assert.True(t.T(), strings.Contains(err.Error(), "context canceled"))
}

func (t *Test_FormRequest_when_calling_submit) Test_multipart_form_successfully_submitted() {
	ret := t.getTestServerAndClient()
	defer ret.Server.Close()

	ret.Mux.HandleFunc("/resource/test", func(rw http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(4096)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			assert.Fail(t.T(), "error parsing form", err.Error())
			return
		}

		if r.Form.Get("foo") != "test" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			assert.Fail(t.T(), "form field foo not supplied", "")
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("response"))
	}).Methods("POST")

	ret.Mux.HandleFunc("/resource", func(rw http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(&Resource{
			Forms: map[string]*Form{
				"test": &Form{
					Action:  "/resource/test",
					Method:  POST,
					Enctype: MediaTypeMultipartFormData,
					Type:    "none",
					Fields: []*FormField{
						&FormField{
							Name:     "foo",
							Type:     MediaTypeHMAPIString,
							Required: true,
						},
					},
				},
			},
		})

		if err != nil {
			assert.FailNow(t.T(), "error marshal json", err)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}).Methods("GET")

	resp, err := ret.Client.Resource("/resource").Form("test").AddFieldAsString("foo", "test").Submit(context.Background())

	assert.Nil(t.T(), err)
	assert.NotNil(t.T(), resp)
	assert.Equal(t.T(), http.StatusOK, resp.StatusCode)

	respbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t.T(), "response", string(respbody))
}

func (t *Test_FormRequest_when_calling_submit) getTestServerAndClient() (ret struct {
	Mux    *mux.Router
	Host   string
	Port   int
	Server *httptest.Server
	Client Client
}) {
	mux := mux.NewRouter()
	svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Println("TEST SVR REQUEST", r)
		mux.ServeHTTP(rw, r)
	}))

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

func TestRunFormTestSuites(t *testing.T) {
	suite.Run(t, new(Test_FormRequest_when_calling_submit))
}
