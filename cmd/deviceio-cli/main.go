package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"github.com/deviceio/cli"
	"github.com/deviceio/cli/device/fs"
	"github.com/deviceio/hmapi/hmclient"
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
	deviceID      = deviceCommand.Flag("device-id", "The hostname or ID of a device").Short('d').Required().String()

	fsReadCommand = deviceCommand.Command("fs:read", "read a file from a device to cli stdout")
	fsReadPath    = fsReadCommand.Arg("path", "Path to the file to read").Required().String()

	fsWriteCommand = deviceCommand.Command("fs:write", "write data from cli stdin to file on device")
	fsWritePath    = fsWriteCommand.Arg("path", "Path to the file to write").Required().String()
	fsWriteAppend  = fsWriteCommand.Flag("append", "append data to end of file").Default("false").Bool()

	sysExecCommand = deviceCommand.Command("sys:exec", "execute a shell command on the remote device")
	sysExecCmd     = sysExecCommand.Arg("cmd", "binary or executable file to execute").Required().String()
	sysExecArgs    = sysExecCommand.Arg("args", "arguments").Strings()
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

	case fsReadCommand.FullCommand():
		loadConfig()
		fs.Read(*deviceID, *fsReadPath, createClient())

	case fsWriteCommand.FullCommand():
		loadConfig()
		fs.Write(*deviceID, *fsWritePath, *fsWriteAppend, createClient())
	}
}

func loadConfig() {
	if err := viper.ReadInConfig(); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Panic("Error loading profile configuration. Please run configure.")
	}
}

func createClient() hmclient.Client {
	return hmclient.New(
		hmclient.SchemeHTTPS,
		viper.GetString("hub_api_addr"),
		viper.GetInt("hub_api_port"),
		&cli.HMAPIAuth{
			UserID:         viper.GetString("user_id"),
			UserTOTPSecret: viper.GetString("user_totp_secret"),
			UserPrivateKey: viper.GetString("user_private_key"),
		},
	)
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
