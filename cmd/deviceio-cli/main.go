package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/Songmu/prompter"
	"github.com/alecthomas/kingpin"
	"github.com/deviceio/cli/device/fs"
	"github.com/deviceio/cli/device/sys"
	"github.com/deviceio/cli/hub"
	"github.com/deviceio/dsc"
	"github.com/deviceio/hmapi"
	sdk "github.com/deviceio/sdk/go-sdk"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/palantir/stacktrace"
	"github.com/spf13/viper"
)

type cliconfig struct {
	HubAddr        string `json:"hub_api_addr,omitempty"`
	HubPort        int    `json:"hub_api_port,omitempty"`
	TLSSkipVerify  bool   `json:"hub_api_skip_cert_verify,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	UserTOTPSecret string `json:"user_totp_secret,omitempty"`
	UserPrivateKey string `json:"user_private_key,omitempty"`
}

var (
	cliApp     = kingpin.New("cli", "Deviceio Command Line Interface")
	cliProfile = cliApp.Flag("profile", "configuration profile to use. default is 'default'").Default("default").String()

	configCommand = cliApp.Command("configure", "Configure deviceio-cli")

	deviceCommand = cliApp.Command("device", "invoke device functionality")

	deviceFSReadCommand = deviceCommand.Command("fs:read", "read a file from a device to cli stdout")
	deviceFSReadDevice  = deviceFSReadCommand.Arg("device-id", "id or hostname of the device").Required().String()
	deviceFSReadPath    = deviceFSReadCommand.Arg("path", "Path to the file to read").Required().String()

	deviceFSWriteCommand = deviceCommand.Command("fs:write", "write data from cli stdin to file on device")
	deviceFSWriteDevice  = deviceFSWriteCommand.Arg("device-id", "id or hostname of the device").Required().String()
	deviceFSWritePath    = deviceFSWriteCommand.Arg("path", "Path to the file to write").Required().String()
	deviceFSWriteAppend  = deviceFSWriteCommand.Flag("append", "append data to end of file").Default("false").Bool()

	deviceExecCommand = deviceCommand.Command("exec", "execute a shell command on the remote device")
	deviceExecDevice  = deviceExecCommand.Arg("device-id", "id or hostname of the device").Required().String()
	deviceExecCmd     = deviceExecCommand.Arg("cmd", "binary or executable file to execute").Required().String()
	deviceExecArgs    = deviceExecCommand.Arg("args", "arguments. If arguments contain a hyphen (-) you must specify (--) before the arg to ignore flag parsing from that point forward").Strings()

	hubCommand = cliApp.Command("hub", "invoke hub functionality")

	hubProxyCommand = hubCommand.Command("proxy", "hosts a local http proxy that signs requests to the hub api")
	hubProxyPort    = hubProxyCommand.Flag("port", "The local port to listen on for http connections").Required().Int()
)

func main() {
	homedir, err := homedir.Dir()

	if err != nil {
		log.Fatal(stacktrace.Propagate(err, "unable to locate user home directory"))
	}

	cliParse := kingpin.MustParse(cliApp.Parse(os.Args[1:]))
	homePath := strings.Replace(fmt.Sprintf("%v/.deviceio/cli/", homedir), "\\", "/", -1)
	configPath := fmt.Sprintf("%v/%v.json", homePath, *cliProfile)

	ensureProfileConfigExists(configPath)

	viper.SetConfigName(*cliProfile)
	viper.AddConfigPath(homePath)
	viper.AddConfigPath("$HOME/.deviceio/cli/")
	viper.AddConfigPath(".")
	viper.SetDefault("hub_api_addr", "127.0.0.1")
	viper.SetDefault("hub_api_port", 4431)
	viper.SetDefault("hub_api_skip_cert_verify", false)
	viper.SetDefault("user_id", "")
	viper.SetDefault("user_totp_secret", "")
	viper.SetDefault("user_private_key", "")

	switch cliParse {
	case configCommand.FullCommand():
		loadConfig()
		configure()

	case deviceFSReadCommand.FullCommand():
		loadConfig()
		fs.Read(*deviceFSReadDevice, *deviceFSReadPath, createClient())

	case deviceFSWriteCommand.FullCommand():
		loadConfig()
		fs.Write(*deviceFSWriteDevice, *deviceFSWritePath, *deviceFSWriteAppend, createClient())

	case deviceExecCommand.FullCommand():
		loadConfig()
		sys.Exec(*deviceExecDevice, *deviceExecCmd, *deviceExecArgs, createSDKClient())

	case hubProxyCommand.FullCommand():
		loadConfig()
		hub.Proxy(viper.GetString("hub_api_addr"), viper.GetInt("hub_api_port"), *hubProxyPort, &sdk.ClientAuth{
			UserID:         viper.GetString("user_id"),
			UserTOTPSecret: viper.GetString("user_totp_secret"),
			UserPrivateKey: viper.GetString("user_private_key"),
		})
	}
}

func ensureProfileConfigExists(configPath string) {
	f := &dsc.File{
		Path:   configPath,
		Absent: false,
	}

	if _, err := f.Apply(); err != nil {
		log.Fatal(stacktrace.Propagate(
			err,
			"failed to create profile configuration file",
		))
	}

	if content, err := ioutil.ReadFile(configPath); err != nil {
		log.Fatal(stacktrace.Propagate(
			err,
			"failed to write default profile configuration file json",
		))
	} else {
		if string(content) == "" {
			ioutil.WriteFile(configPath, []byte("{}"), 0666)
		}
	}
}

func loadConfig() {
	if err := viper.ReadInConfig(); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Panic("Error loading profile configuration. Please run configure.")
	}
}

func createClient() hmapi.Client {
	return hmapi.NewClient(&hmapi.ClientConfig{
		Auth: &sdk.ClientAuth{
			UserID:         viper.GetString("user_id"),
			UserTOTPSecret: viper.GetString("user_totp_secret"),
			UserPrivateKey: viper.GetString("user_private_key"),
		},
		Scheme: hmapi.HTTPS,
		Host:   viper.GetString("hub_api_addr"),
		Port:   viper.GetInt("hub_api_port"),
	})
}

func createSDKClient() sdk.Client {
	return sdk.NewClient(sdk.ClientConfig{
		UserID:     viper.GetString("user_id"),
		TOTPSecret: viper.GetString("user_totp_secret"),
		PrivateKey: viper.GetString("user_private_key"),
		HubHost:    viper.GetString("hub_api_addr"),
		HubPort:    viper.GetInt("hub_api_port"),
	})
}

func configure() {
	homedir, err := homedir.Dir()

	if err != nil {
		panic(err)
	}

	answers := &cliconfig{
		HubAddr: prompter.Prompt("Hub API Address or Hostname", viper.GetString("hub_api_addr")),
		HubPort: func() int {
			port := prompter.Prompt("Hub API Port", viper.GetString("hub_api_port"))

			i, err := strconv.Atoi(port)

			if err != nil {
				log.Fatal(err)
			}

			return i
		}(),
		UserID:         prompter.Prompt("User ID", viper.GetString("user_id")),
		UserPrivateKey: prompter.Password("User Private Key"),
		UserTOTPSecret: prompter.Password("User TOTP Secret"),
	}

	jsonb, err := json.MarshalIndent(answers, "", "    ")

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
