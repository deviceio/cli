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

func Write(deviceid, path string, append bool, c hmapi.Client) {
	datar, dataw := io.Pipe()
	done := make(chan bool)

	go func() {
		defer dataw.Close()
		defer datar.Close()

		buf := make([]byte, 250000)

		for {
			nr, rerr := os.Stdin.Read(buf)

			if rerr != nil && rerr != io.EOF {
				logrus.Fatal(rerr)
			}

			if rerr != nil && rerr == io.EOF {
				break
			}

			if nr > 0 || rerr == io.EOF {
				dataw.Write(buf[:nr])
				continue
			}

			break
		}
	}()

	go func() {
		formResult, err := c.
			Resource(fmt.Sprintf("/device/%v/filesystem", deviceid)).
			Form("write").
			AddFieldAsString("path", path).
			AddFieldAsBool("append", append).
			AddFieldAsOctetStream("data", datar).
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

		close(done)
	}()

	<-done
}
