package fs

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/hmapi"
)

func Read(deviceid, path string, c hmapi.Client) {
	formResult, err := c.
		Resource(fmt.Sprintf("/device/%v/filesystem", deviceid)).
		Form("read").
		AddFieldAsString("path", path).
		Submit(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	resp := formResult

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
