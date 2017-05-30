package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/hmapi/hmclient"
)

func Read(deviceid, path string, c hmclient.Client) {
	formResult, err := c.
		Resource(fmt.Sprintf("/device/%v/filesystem", deviceid)).
		Form("read").
		SetFieldAsString("path", path).
		Submit()

	if err != nil {
		log.Fatal(err)
	}

	resp := formResult.RawResponse()

	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"endpoint":     resp.Request.URL.Path,
			"method":       resp.Request.Method,
			"statusCode":   resp.StatusCode,
			"responseBody": string(body),
		}).Fatal("Error calling device endpoint")
	}

	defer func() {
		if resp.Trailer.Get("Error") != "" {
			os.Stderr.Write([]byte(resp.Trailer.Get("Error")))
			os.Stderr.Sync()
		}
	}()

	buf := make([]byte, 250000)

	if _, err := io.CopyBuffer(os.Stdout, resp.Body, buf); err != nil {
		os.Stderr.Write([]byte(err.Error()))
	}
}
