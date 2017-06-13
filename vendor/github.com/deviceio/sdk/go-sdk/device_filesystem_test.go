package sdk

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"net/http"

	"strings"

	"github.com/deviceio/agent/resources/filesystem"
	"github.com/deviceio/hmapi"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Test_DeviceFilesystem struct {
	suite.Suite
}

func (t *Test_DeviceFilesystem) Test_successfull_read() {
	objects := t.getTestObjects()
	defer objects.server.Close()

	fsroot := &filesystem.Root{}

	objects.mux.HandleFunc("/device/{id}/filesystem", fsroot.Get)
	objects.mux.HandleFunc("/filesystem/read", fsroot.Read)

	tmpfile, err := ioutil.TempFile("", "go-sdk-test")

	if err != nil {
		t.T().Error("Error creating temp file", err.Error())
		t.T().FailNow()
	}

	defer tmpfile.Close()
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	ioutil.WriteFile(tmpfile.Name(), []byte("hello"), 0666)

	reader := objects.client.Device("whatever").Filesystem().Reader(
		context.Background(),
		tmpfile.Name(),
		0,
		-1,
	)

	filedata, err := ioutil.ReadAll(reader)

	assert.NotNil(t.T(), filedata)
	assert.Nil(t.T(), err)
	assert.Equal(t.T(), "hello", string(filedata))
}

func (t *Test_DeviceFilesystem) Test_successfull_write() {
	objects := t.getTestObjects()
	defer objects.server.Close()

	fsroot := &filesystem.Root{}

	objects.mux.HandleFunc("/device/{id}/filesystem", fsroot.Get)
	objects.mux.HandleFunc("/filesystem/write", fsroot.Write)

	tmpfile, err := ioutil.TempFile("", "go-sdk-device-filesystem-writer")

	if err != nil {
		t.T().Error("Error creating temp file", err.Error())
		t.T().FailNow()
	}

	tmpfile.Close()
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	writer := objects.client.Device("/whatever").Filesystem().Writer(
		context.Background(),
		tmpfile.Name(),
		false,
	)
	defer writer.Close()

	data := strings.NewReader("hello")

	nw, err := io.Copy(writer, data)

	assert.Equal(t.T(), int64(5), nw)
	assert.Nil(t.T(), err)

	<-time.After(250 * time.Millisecond) //allow some time for the fs to catch up

	filedata, err := ioutil.ReadFile(tmpfile.Name())

	assert.Equal(t.T(), "hello", string(filedata))
	assert.Nil(t.T(), err)
}

func (t *Test_DeviceFilesystem) getTestObjects() (objects struct {
	server *httptest.Server
	client Client
	mux    *mux.Router
}) {
	mux := mux.NewRouter()
	svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Println("TEST SVR REQUEST", r)
		mux.ServeHTTP(rw, r)
	}))

	url, _ := url.Parse(svr.URL)

	hoststr, portstr, _ := net.SplitHostPort(url.Host)
	port, _ := strconv.ParseInt(portstr, 10, 0)

	client := NewClient(ClientConfig{
		HMClient: hmapi.NewClient(&hmapi.ClientConfig{
			Auth:   &hmapi.AuthNone{},
			Host:   hoststr,
			Port:   int(port),
			Scheme: hmapi.HTTP,
		}),
	})

	objects.mux = mux
	objects.server = svr
	objects.client = client

	return
}

func TestDeviceFilesystemSuite(t *testing.T) {
	suite.Run(t, new(Test_DeviceFilesystem))
}
