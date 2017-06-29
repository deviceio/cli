package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"github.com/deviceio/cli/device/fs"
	"github.com/deviceio/cli/device/sys"
	"github.com/deviceio/cli/hub"
	"github.com/deviceio/hmapi"
	sdk "github.com/deviceio/sdk/go-sdk"
	homedir "github.com/mitchellh/go-homedir"
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

	configCommand        = cliApp.Command("configure", "Configures deviceio-cli")
	configHubAddr        = configCommand.Arg("hub-api-address", "Your hub api ip or hostname").Required().String()
	configHubPort        = configCommand.Arg("hub-api-port", "The port to access the hub api on").Required().Int()
	configUserID         = configCommand.Arg("user-id", "Your user email, login or uuid").Required().String()
	configUserTOTPSecret = configCommand.Arg("user-totp-secret", "Your user totp secret").Required().String()
	configUserPrivateKey = configCommand.Arg("user-private-key", "Your user private key").Required().String()
	configSkipTLSVerify  = configCommand.Flag("insecure", "Do not verify hub api tls certificate").Short('i').Bool()

	deviceCommand = cliApp.Command("device", "invoke device functionality")
	//deviceID      = deviceCommand.Arg("device-id", "The hostname or ID of a device").Required().String()

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
		logrus.Panic(err)
	}

	cliParse := kingpin.MustParse(cliApp.Parse(os.Args[1:]))

	viper.SetConfigName(*cliProfile)
	viper.AddConfigPath(strings.Replace(fmt.Sprintf("%v/.deviceio/cli/", homedir), "\\", "/", -1))
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
		config(*configHubAddr, *configHubPort, *configUserID, *configUserTOTPSecret, *configUserPrivateKey, *configSkipTLSVerify)

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

func config(hubAddr string, hubPort int, userID string, userTOTPSecret string, userPrivateKey string, skipTLSVerify bool) {
	homedir, err := homedir.Dir()

	if err != nil {
		panic(err)
	}

	jsonb, err := json.MarshalIndent(&cliconfig{
		HubAddr:        hubAddr,
		HubPort:        hubPort,
		UserID:         userID,
		UserTOTPSecret: userTOTPSecret,
		UserPrivateKey: userPrivateKey,
		TLSSkipVerify:  skipTLSVerify,
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
