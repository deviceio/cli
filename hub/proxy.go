package hub

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	sdk "github.com/deviceio/sdk/go-sdk"
	"github.com/deviceio/shared/types"
)

type bufpool struct {
	size int64
}

func (t *bufpool) Get() []byte {
	return make([]byte, t.size)
}

func (t *bufpool) Put([]byte) {
}

func Proxy(hubHost string, hubPort int, localPort int, c *sdk.ClientAuth) {
	certpath, keypath := makeTempCertificates()

	rpurl, err := url.Parse(fmt.Sprintf("https://%v:%v/", hubHost, hubPort))

	if err != nil {
		logrus.WithField("error", err.Error()).Fatal("Error parsing reverse proxy url")
	}

	rp := types.NewSingleHostReverseProxy(rpurl)
	rpdir := rp.Director
	rp.Director = func(r *http.Request) {
		rpdir(r)
		c.Sign(r)
	}

	rp.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	rp.BufferPool = &bufpool{
		size: 250000,
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(rw, r)
	})

	logrus.WithFields(logrus.Fields{
		"port":     localPort,
		"tempcert": certpath,
		"tempkey":  keypath,
	}).Info("Starting local hub api proxy")

	http.ListenAndServeTLS(fmt.Sprintf(":%v", localPort), certpath, keypath, nil)
}

func makeTempCertificates() (string, string) {
	certgen := &types.CertGen{
		Host:      "localhost",
		ValidFrom: "Jan 1 15:04:05 2011",
		ValidFor:  8760 * time.Hour,
		IsCA:      false,
		RsaBits:   4096,
	}

	var err error
	var certBytes []byte
	var certTemp *os.File
	var keyBytes []byte
	var keyTemp *os.File

	certBytes, keyBytes = certgen.Generate()

	if certTemp, err = ioutil.TempFile("", "deviceio-cli"); err != nil {
		logrus.Fatal(err.Error())
	}
	defer certTemp.Close()

	if keyTemp, err = ioutil.TempFile("", "deviceio-cli"); err != nil {
		logrus.Fatal(err.Error())
	}
	defer keyTemp.Close()

	io.Copy(certTemp, bytes.NewBuffer(certBytes))
	io.Copy(keyTemp, bytes.NewBuffer(keyBytes))

	return certTemp.Name(), keyTemp.Name()
}
