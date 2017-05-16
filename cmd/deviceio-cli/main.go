// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"github.com/deviceio/hmapi/hmclient"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/viper"
)

type cliconfig struct {
	HubAddr       string `json:"hub_api_addr,omitempty"`
	HubPort       int    `json:"hub_api_port,omitempty"`
	TLSSkipVerify bool   `json:"hub_api_skip_cert_verify,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	UserPassword  string `json:"user_password,omitempty"`
	UserSecret    string `json:"user_secret,omitempty"`
}

type hmapiAuth struct {
	userID       string
	userPassword string
	userSecret   string
}

func (t *hmapiAuth) Sign(r *http.Request) {
	code, _ := totp.GenerateCode(t.userSecret, time.Now())
	r.Header.Set("Authorization", fmt.Sprintf("totp %v:%v:%v", t.userID, t.userPassword, code))
}

var (
	cli        = kingpin.New("cli", "Deviceio Command Line Interface")
	cliProfile = cli.Flag("profile", "configuration profile to use. default is 'default'").Default("default").String()

	configCommand       = cli.Command("configure", "Configures deviceio-cli")
	configHubAddr       = configCommand.Arg("hub-api-address", "Your hub api ip or hostname").Required().String()
	configHubPort       = configCommand.Arg("hub-api-port", "The port to access the hub api on").Required().Int()
	configUserID        = configCommand.Arg("user-id", "Your user email, login or uuid").Required().String()
	configUserPassword  = configCommand.Arg("user-password", "Your user password").Required().String()
	configUserSecret    = configCommand.Arg("user-secret", "Your user secret for generating totp passcodes").Required().String()
	configSkipTLSVerify = configCommand.Flag("insecure", "Do not verify hub api tls certificate").Short('i').Bool()

	readCommand  = cli.Command("read", "read a file from a device to cli stdout")
	readDeviceID = readCommand.Arg("deviceid", "The device id or hostname").Required().String()
	readPath     = readCommand.Arg("path", "Path to the file to read").Required().String()

	writeCommand  = cli.Command("write", "write data from cli stdin to file on device")
	writeDeviceID = writeCommand.Arg("deviceid", "The device id or hostname").Required().String()
	writePath     = writeCommand.Arg("path", "Path to the file to write").Required().String()
	writeAppend   = writeCommand.Flag("append", "append data to end of file").Default("false").Bool()
)

func main() {
	homedir, err := homedir.Dir()

	if err != nil {
		logrus.Panic(err)
	}

	cliParse := kingpin.MustParse(cli.Parse(os.Args[1:]))

	viper.SetConfigName(*cliProfile)
	viper.AddConfigPath(strings.Replace(fmt.Sprintf("%v/.deviceio/cli/", homedir), "\\", "/", -1))
	viper.AddConfigPath("$HOME/.deviceio/cli/")
	viper.AddConfigPath(".")

	viper.SetDefault("hub_api_addr", "127.0.0.1")
	viper.SetDefault("hub_api_port", 4431)
	viper.SetDefault("hub_api_skip_cert_verify", false)
	viper.SetDefault("user_id", "")
	viper.SetDefault("user_password", "")
	viper.SetDefault("user_secret", "")

	switch cliParse {
	case configCommand.FullCommand():
		config(*configHubAddr, *configHubPort, *configUserID, *configUserPassword, *configUserSecret, *configSkipTLSVerify)

	case readCommand.FullCommand():
		loadConfig()
		read(*readDeviceID, *readPath)

	case writeCommand.FullCommand():
		loadConfig()
		write(*writeDeviceID, *writePath, *writeAppend)
	}
}

func loadConfig() {
	if err := viper.ReadInConfig(); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Panic("Error loading profile configuration. Please run configure.")
	}
}

func config(hubAddr string, hubPort int, userID string, userPassword string, userSecret string, skipTLSVerify bool) {
	homedir, err := homedir.Dir()

	if err != nil {
		panic(err)
	}

	jsonb, err := json.MarshalIndent(&cliconfig{
		HubAddr:       hubAddr,
		HubPort:       hubPort,
		UserID:        userID,
		UserPassword:  userPassword,
		UserSecret:    userSecret,
		TLSSkipVerify: skipTLSVerify,
	}, "", "    ")

	if err != nil {
		panic(err)
	}

	cfgdir := fmt.Sprintf("%v/.deviceio/cli", homedir)
	cfgfile := fmt.Sprintf("%v/%v.json", cfgdir, *cliProfile)

	if err := os.MkdirAll(cfgdir, 0666); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(cfgfile, jsonb, 0666); err != nil {
		panic(err)
	}
}

func read(deviceid, path string) {
	c := hmclient.New(
		hmclient.SchemeHTTPS,
		viper.GetString("hub_api_addr"),
		viper.GetInt("hub_api_port"),
		&hmapiAuth{
			userID:       viper.GetString("user_id"),
			userPassword: viper.GetString("user_password"),
			userSecret:   viper.GetString("user_secret"),
		},
	)

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

func write(deviceid, path string, append bool) {
	c := hmclient.New(
		hmclient.SchemeHTTPS,
		viper.GetString("hub_api_addr"),
		viper.GetInt("hub_api_port"),
		&hmapiAuth{
			userID:       viper.GetString("user_id"),
			userPassword: viper.GetString("user_password"),
			userSecret:   viper.GetString("user_secret"),
		},
	)

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
