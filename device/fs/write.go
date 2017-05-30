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

func Write(deviceid, path string, append bool, c hmclient.Client) {
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
			SetFieldAsString("path", path).
			SetFieldAsBool("append", append).
			SetFieldAsOctetStream("data", datar).
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

		close(done)
	}()

	<-done
}
