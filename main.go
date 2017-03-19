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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"io/ioutil"

	"log"

	"github.com/alecthomas/kingpin"
	homedir "github.com/mitchellh/go-homedir"
)

type cliconfig struct {
	HubURL        string
	UserKey       string
	UserSecret    string
	TLSSkipVerify bool
}

var (
	cfg    = &cliconfig{}
	cfgdir = ""
	client *http.Client
)

var (
	cli = kingpin.New("cli", "Deviceio Command Line Interface")

	configCommand       = cli.Command("configure", "Configure deviceio-cli")
	configHubURL        = configCommand.Arg("hub-url", "Your hub api url").Required().String()
	configUserKey       = configCommand.Arg("user-key", "Your user key").Required().String()
	configUserSecret    = configCommand.Arg("user-secret", "Your user secret").Required().String()
	configSkipTLSVerify = configCommand.Flag("insecure", "Do not verify hub api tls certificate").Short('i').Bool()

	catCommand  = cli.Command("cat", "read a file from a device")
	catDeviceID = catCommand.Arg("deviceid", "The target device").Required().String()
	catPath     = catCommand.Arg("path", "Path to the file to read").Required().String()
)

func main() {
	mkconfigdir()

	switch kingpin.MustParse(cli.Parse(os.Args[1:])) {
	case configCommand.FullCommand():
		config(*configHubURL, *configUserKey, *configUserSecret, *configSkipTLSVerify)

	case catCommand.FullCommand():
		cat(*catDeviceID, *catPath)
	}
}

func mkconfigdir() {
	dir, err := homedir.Dir()

	if err != nil {
		panic(err)
	}

	dir = fmt.Sprintf("%v/.deviceio/cli", dir)
	dir = strings.Replace(dir, "\\", "/", -1)

	err = os.MkdirAll(dir, 0700)

	if err != nil {
		panic(err)
	}

	cfgdir = dir
}

func ldconfig() {
	file, err := os.Open(fmt.Sprintf("%v/default.json", cfgdir))
	defer file.Close()

	if os.IsNotExist(err) {
		panic("No configuration found. Please run 'configure'")
	}

	err = json.NewDecoder(file).Decode(&cfg)

	if err != nil {
		panic(err)
	}

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.TLSSkipVerify,
			},
		},
	}
}

func config(hubURL, userKey, userSecret string, tlsSkipVerify bool) {
	jsonb, err := json.MarshalIndent(&cliconfig{
		HubURL:        hubURL,
		UserKey:       userKey,
		UserSecret:    userSecret,
		TLSSkipVerify: tlsSkipVerify,
	}, "", "    ")

	if err != nil {
		panic(err)
	}

	defaultcfgf := fmt.Sprintf("%v/default.json", cfgdir)

	err = ioutil.WriteFile(defaultcfgf, jsonb, 0700)

	if err != nil {
		panic(err)
	}
}

func cat(deviceid, path string) {
	ldconfig()

	if deviceid == "" {
		panic("Device id not specified")
	}

	r, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%v/device/%v/filesystem/read", strings.TrimRight(cfg.HubURL, "/"), deviceid),
		nil,
	)

	r.Header.Add("X-Path", path)

	if err != nil {
		panic(err)
	}

	resp, err := client.Do(r)

	if err != nil {
		log.Fatal(resp, err)
	}

	if resp.StatusCode >= 300 {
		log.Fatal(resp, err)
	}

	log.Println(
		resp.TransferEncoding,
		resp.Header,
	)

	buf := make([]byte, 250000)

	io.CopyBuffer(os.Stdout, resp.Body, buf)
}
